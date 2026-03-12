package whatsapp

import (
	"context"
	"fmt"
	"time"

	waEvents "go.mau.fi/whatsmeow/types/events"
	waTypes "go.mau.fi/whatsmeow/types"
)

func (i *Instance) handleEvent(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *waEvents.Connected:
		i.handleConnected()
	case *waEvents.Disconnected:
		i.handleDisconnected()
	case *waEvents.LoggedOut:
		i.handleLoggedOut(evt)
	case *waEvents.Message:
		i.handleMessage(evt)
	case *waEvents.Receipt:
		i.handleReceipt(evt)
	case *waEvents.Presence:
		i.handlePresence(evt)
	case *waEvents.ChatPresence:
		i.handleChatPresence(evt)
	case *waEvents.GroupInfo:
		i.handleGroupInfo(evt)
	case *waEvents.JoinedGroup:
		i.handleJoinedGroup(evt)
	case *waEvents.CallOffer:
		i.handleCallOffer(evt)
	case *waEvents.CallOfferNotice:
		i.handleCallOfferNotice(evt)
	case *waEvents.PushName:
		i.handlePushName(evt)
	case *waEvents.KeepAliveTimeout:
		i.logger.Warn("keepalive timeout", "error_count", evt.ErrorCount)
	case *waEvents.KeepAliveRestored:
		i.logger.Info("keepalive restored")
	}
}

func (i *Instance) handleConnected() {
	i.connected.Store(true)
	i.logger.Info("connected to WhatsApp")

	// Update status and phone info
	ctx := context.Background()
	if i.client.Store.ID != nil {
		jid := i.client.Store.ID.String()
		phone := i.client.Store.ID.User
		_ = i.dbStore.UpdateInstanceStatus(ctx, i.dbInst.ID, string(StateOpen))
		_ = i.dbStore.UpdateInstancePhone(ctx, i.dbInst.ID, phone, i.client.Store.BusinessName, jid)
		i.mu.Lock()
		i.dbInst.Status = string(StateOpen)
		i.dbInst.Phone = phone
		i.dbInst.WhatsmeowJID = jid
		i.mu.Unlock()
	}

	// Apply always_online setting
	settings := i.getSettings()
	if settings != nil && settings.AlwaysOnline {
		_ = i.client.SendPresence(context.Background(), waTypes.PresenceAvailable)
	}

	i.emit(EventConnectionUpdate, ConnectionUpdate{State: string(StateOpen)})
}

func (i *Instance) handleDisconnected() {
	i.connected.Store(false)
	i.logger.Info("disconnected from WhatsApp")
	i.emit(EventConnectionUpdate, ConnectionUpdate{State: string(StateClose), Reason: "disconnected"})
}

func (i *Instance) handleLoggedOut(evt *waEvents.LoggedOut) {
	i.connected.Store(false)
	i.logger.Info("logged out from WhatsApp", "reason", evt.Reason)

	ctx := context.Background()
	_ = i.dbStore.UpdateInstanceStatus(ctx, i.dbInst.ID, string(StateClose))
	_ = i.dbStore.UpdateInstancePhone(ctx, i.dbInst.ID, "", "", "")

	i.emit(EventConnectionUpdate, ConnectionUpdate{
		State:  string(StateClose),
		Reason: "logged_out",
	})
}

func (i *Instance) handleMessage(evt *waEvents.Message) {
	settings := i.getSettings()

	// Skip group messages if groups_ignore is enabled
	if settings != nil && settings.GroupsIgnore && evt.Info.IsGroup {
		return
	}

	// Auto-read messages
	if settings != nil && settings.ReadMessages && !evt.Info.IsFromMe {
		_ = i.client.MarkRead(
			context.Background(),
			[]waTypes.MessageID{evt.Info.ID},
			time.Now(),
			evt.Info.Chat,
			evt.Info.Sender,
		)
	}

	contentType, content := extractMessageContent(evt)

	payload := MessagePayload{
		MessageID:   string(evt.Info.ID),
		ChatJID:     evt.Info.Chat.String(),
		SenderJID:   evt.Info.Sender.String(),
		SenderName:  evt.Info.PushName,
		IsGroup:     evt.Info.IsGroup,
		IsFromMe:    evt.Info.IsFromMe,
		ContentType: contentType,
		Content:     content,
		Timestamp:   evt.Info.Timestamp,
	}

	if evt.Info.IsFromMe {
		i.emit(EventMessageSent, payload)
	} else {
		i.emit(EventMessageReceived, payload)
	}
}

func (i *Instance) handleReceipt(evt *waEvents.Receipt) {
	var receiptType string
	switch evt.Type {
	case waTypes.ReceiptTypeRead, waTypes.ReceiptTypeReadSelf:
		receiptType = "read"
	case waTypes.ReceiptTypePlayed:
		receiptType = "played"
	case waTypes.ReceiptTypeDelivered:
		receiptType = "delivered"
	default:
		return
	}

	msgIDs := make([]string, len(evt.MessageIDs))
	for idx, id := range evt.MessageIDs {
		msgIDs[idx] = string(id)
	}

	i.emit(EventMessageUpdated, ReceiptPayload{
		MessageIDs: msgIDs,
		ChatJID:    evt.MessageSource.Chat.String(),
		SenderJID:  evt.MessageSource.Sender.String(),
		Type:       receiptType,
		Timestamp:  evt.Timestamp,
	})
}

func (i *Instance) handlePresence(evt *waEvents.Presence) {
	var lastSeen *time.Time
	if !evt.LastSeen.IsZero() {
		lastSeen = &evt.LastSeen
	}

	i.emit(EventPresenceUpdate, PresencePayload{
		JID:       evt.From.String(),
		Available: !evt.Unavailable,
		LastSeen:  lastSeen,
	})
}

func (i *Instance) handleChatPresence(evt *waEvents.ChatPresence) {
	i.emit(EventPresenceUpdate, PresencePayload{
		JID:          evt.MessageSource.Sender.String(),
		ChatPresence: string(evt.State),
	})
}

func (i *Instance) handleGroupInfo(evt *waEvents.GroupInfo) {
	payload := GroupUpdatePayload{
		JID:       evt.JID.String(),
		SenderJID: evt.Sender.String(),
	}

	if evt.Name != nil {
		payload.Name = evt.Name.Name
	}
	if evt.Topic != nil {
		payload.Topic = evt.Topic.Topic
	}

	// Handle participant changes
	if len(evt.Join) > 0 {
		jids := jidsToStrings(evt.Join)
		i.emit(EventGroupParticipants, GroupParticipantsPayload{
			JID: evt.JID.String(), Action: "join", Participants: jids, SenderJID: evt.Sender.String(),
		})
	}
	if len(evt.Leave) > 0 {
		jids := jidsToStrings(evt.Leave)
		i.emit(EventGroupParticipants, GroupParticipantsPayload{
			JID: evt.JID.String(), Action: "leave", Participants: jids, SenderJID: evt.Sender.String(),
		})
	}
	if len(evt.Promote) > 0 {
		jids := jidsToStrings(evt.Promote)
		i.emit(EventGroupParticipants, GroupParticipantsPayload{
			JID: evt.JID.String(), Action: "promote", Participants: jids, SenderJID: evt.Sender.String(),
		})
	}
	if len(evt.Demote) > 0 {
		jids := jidsToStrings(evt.Demote)
		i.emit(EventGroupParticipants, GroupParticipantsPayload{
			JID: evt.JID.String(), Action: "demote", Participants: jids, SenderJID: evt.Sender.String(),
		})
	}

	if payload.Name != "" || payload.Topic != "" {
		i.emit(EventGroupUpdate, payload)
	}
}

func (i *Instance) handleJoinedGroup(evt *waEvents.JoinedGroup) {
	i.emit(EventGroupUpdate, GroupUpdatePayload{
		JID:  evt.JID.String(),
		Name: evt.GroupName.Name,
	})
}

func (i *Instance) handleCallOffer(evt *waEvents.CallOffer) {
	settings := i.getSettings()
	if settings != nil && settings.RejectCall {
		// Reject the call (whatsmeow does not have a built-in reject method,
		// but we can send a text message if configured)
		if settings.RejectCallMessage != "" && i.client != nil {
			go func() {
				_, _ = i.SendText(context.Background(), evt.CallCreator.String(), settings.RejectCallMessage)
			}()
		}
	}

	i.emit(EventCallReceived, CallPayload{
		CallerJID: evt.CallCreator.String(),
		CallID:    evt.CallID,
		Type:      "audio",
	})
}

func (i *Instance) handleCallOfferNotice(evt *waEvents.CallOfferNotice) {
	callType := "audio"
	if evt.Media == "video" {
		callType = "video"
	}

	i.emit(EventCallReceived, CallPayload{
		CallerJID: evt.CallCreator.String(),
		CallID:    evt.CallID,
		Type:      callType,
		IsGroup:   evt.Type == "group",
	})
}

func (i *Instance) handlePushName(evt *waEvents.PushName) {
	i.logger.Debug("push name updated", "jid", evt.JID.String(), "name", evt.NewPushName)
}

// extractMessageContent extracts the content type and content from a message event.
func extractMessageContent(evt *waEvents.Message) (string, any) {
	msg := evt.Message

	if msg.GetConversation() != "" {
		return "text", map[string]string{"text": msg.GetConversation()}
	}
	if msg.GetExtendedTextMessage() != nil {
		return "text", map[string]string{
			"text":    msg.GetExtendedTextMessage().GetText(),
			"context": msg.GetExtendedTextMessage().GetContextInfo().String(),
		}
	}
	if msg.GetImageMessage() != nil {
		return "image", map[string]string{
			"caption":  msg.GetImageMessage().GetCaption(),
			"mimetype": msg.GetImageMessage().GetMimetype(),
			"url":      msg.GetImageMessage().GetURL(),
		}
	}
	if msg.GetVideoMessage() != nil {
		return "video", map[string]string{
			"caption":  msg.GetVideoMessage().GetCaption(),
			"mimetype": msg.GetVideoMessage().GetMimetype(),
			"url":      msg.GetVideoMessage().GetURL(),
		}
	}
	if msg.GetAudioMessage() != nil {
		return "audio", map[string]string{
			"mimetype": msg.GetAudioMessage().GetMimetype(),
			"url":      msg.GetAudioMessage().GetURL(),
			"ptt":      fmt.Sprintf("%v", msg.GetAudioMessage().GetPTT()),
		}
	}
	if msg.GetDocumentMessage() != nil {
		return "document", map[string]string{
			"caption":  msg.GetDocumentMessage().GetCaption(),
			"filename": msg.GetDocumentMessage().GetFileName(),
			"mimetype": msg.GetDocumentMessage().GetMimetype(),
			"url":      msg.GetDocumentMessage().GetURL(),
		}
	}
	if msg.GetStickerMessage() != nil {
		return "sticker", map[string]string{
			"mimetype": msg.GetStickerMessage().GetMimetype(),
			"url":      msg.GetStickerMessage().GetURL(),
		}
	}
	if msg.GetLocationMessage() != nil {
		return "location", map[string]any{
			"latitude":  msg.GetLocationMessage().GetDegreesLatitude(),
			"longitude": msg.GetLocationMessage().GetDegreesLongitude(),
			"name":      msg.GetLocationMessage().GetName(),
			"address":   msg.GetLocationMessage().GetAddress(),
		}
	}
	if msg.GetContactMessage() != nil {
		return "contact", map[string]string{
			"display_name": msg.GetContactMessage().GetDisplayName(),
			"vcard":        msg.GetContactMessage().GetVcard(),
		}
	}
	if msg.GetReactionMessage() != nil {
		return "reaction", map[string]string{
			"text":       msg.GetReactionMessage().GetText(),
			"message_id": string(msg.GetReactionMessage().GetKey().GetID()),
		}
	}

	return "unknown", nil
}

func jidsToStrings(jids []waTypes.JID) []string {
	result := make([]string, len(jids))
	for i, j := range jids {
		result[i] = j.String()
	}
	return result
}
