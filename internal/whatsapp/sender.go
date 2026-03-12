package whatsapp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	waEvents "go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// SendText sends a text message.
func (i *Instance) SendText(ctx context.Context, to string, text string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	resp, err := i.client.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending text: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// SendMedia sends a media message (image, video, document, audio).
func (i *Instance) SendMedia(ctx context.Context, to, mediaType, media, caption, filename, mimeType string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	data, err := resolveMediaData(media)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("resolving media: %w", err)
	}

	var waMediaType whatsmeow.MediaType
	switch mediaType {
	case "image":
		waMediaType = whatsmeow.MediaImage
	case "video":
		waMediaType = whatsmeow.MediaVideo
	case "audio":
		waMediaType = whatsmeow.MediaAudio
	case "document":
		waMediaType = whatsmeow.MediaDocument
	default:
		return MessageSendResponse{}, fmt.Errorf("unsupported media type: %s", mediaType)
	}

	uploadResp, err := i.client.Upload(ctx, data, waMediaType)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("uploading media: %w", err)
	}

	msg := buildMediaMessage(mediaType, uploadResp, caption, filename, mimeType, uint64(len(data)))

	resp, err := i.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending media: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// SendLocation sends a location message.
func (i *Instance) SendLocation(ctx context.Context, to string, lat, lon float64, name, address string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	resp, err := i.client.SendMessage(ctx, jid, &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(lat),
			DegreesLongitude: proto.Float64(lon),
			Name:             proto.String(name),
			Address:          proto.String(address),
		},
	})
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending location: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// SendContact sends a vcard contact message.
func (i *Instance) SendContact(ctx context.Context, to, contactName string, phones []string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	vcard := buildVCard(contactName, phones)

	resp, err := i.client.SendMessage(ctx, jid, &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(contactName),
			Vcard:       proto.String(vcard),
		},
	})
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending contact: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// SendReaction sends an emoji reaction to a message.
func (i *Instance) SendReaction(ctx context.Context, to, messageID, emoji string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	msg := i.client.BuildReaction(jid, waTypes.EmptyJID, waTypes.MessageID(messageID), emoji)

	resp, err := i.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending reaction: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// SendSticker sends a sticker message.
func (i *Instance) SendSticker(ctx context.Context, to, stickerData, mimeType string) (MessageSendResponse, error) {
	jid, err := parseJID(to)
	if err != nil {
		return MessageSendResponse{}, err
	}

	data, err := resolveMediaData(stickerData)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("resolving sticker: %w", err)
	}

	if mimeType == "" {
		mimeType = "image/webp"
	}

	uploadResp, err := i.client.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("uploading sticker: %w", err)
	}

	fileLen := uint64(len(data))
	resp, err := i.client.SendMessage(ctx, jid, &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    &fileLen,
			Mimetype:      proto.String(mimeType),
		},
	})
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("sending sticker: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// EditMessage edits a previously sent message.
func (i *Instance) EditMessage(ctx context.Context, chatJID, messageID, text string) (MessageSendResponse, error) {
	jid, err := parseJID(chatJID)
	if err != nil {
		return MessageSendResponse{}, err
	}

	msg := i.client.BuildEdit(jid, waTypes.MessageID(messageID), &waE2E.Message{
		Conversation: proto.String(text),
	})

	resp, err := i.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return MessageSendResponse{}, fmt.Errorf("editing message: %w", err)
	}

	return MessageSendResponse{
		MessageID: string(resp.ID),
		Timestamp: resp.Timestamp,
	}, nil
}

// RevokeMessage deletes a message for everyone.
func (i *Instance) RevokeMessage(ctx context.Context, chatJID, messageID string) error {
	jid, err := parseJID(chatJID)
	if err != nil {
		return err
	}

	msg := i.client.BuildRevoke(jid, waTypes.EmptyJID, waTypes.MessageID(messageID))
	_, err = i.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("revoking message: %w", err)
	}
	return nil
}

// MarkRead marks messages as read.
func (i *Instance) MarkRead(ctx context.Context, chatJID string, messageIDs []string) error {
	jid, err := parseJID(chatJID)
	if err != nil {
		return err
	}

	ids := make([]waTypes.MessageID, len(messageIDs))
	for idx, id := range messageIDs {
		ids[idx] = waTypes.MessageID(id)
	}

	return i.client.MarkRead(ctx, ids, time.Now(), jid, waTypes.EmptyJID)
}

// SendChatPresence sends a typing/recording indicator.
func (i *Instance) SendChatPresence(ctx context.Context, chatJID, presence string) error {
	jid, err := parseJID(chatJID)
	if err != nil {
		return err
	}

	var state waTypes.ChatPresence
	switch presence {
	case "composing":
		state = waTypes.ChatPresenceComposing
	case "paused":
		state = waTypes.ChatPresencePaused
	default:
		state = waTypes.ChatPresenceComposing
	}

	var media waTypes.ChatPresenceMedia
	if presence == "recording" {
		state = waTypes.ChatPresenceComposing
		media = waTypes.ChatPresenceMediaAudio
	}

	return i.client.SendChatPresence(ctx, jid, state, media)
}

// CheckNumbers checks if phone numbers are registered on WhatsApp.
func (i *Instance) CheckNumbers(ctx context.Context, phones []string) ([]NumberCheckResult, error) {
	resp, err := i.client.IsOnWhatsApp(ctx, phones)
	if err != nil {
		return nil, fmt.Errorf("checking numbers: %w", err)
	}

	results := make([]NumberCheckResult, len(resp))
	for idx, r := range resp {
		results[idx] = NumberCheckResult{
			Number: r.Query,
			Exists: r.IsIn,
		}
		if r.IsIn {
			results[idx].JID = r.JID.String()
		}
	}
	return results, nil
}

// UpdateBlocklist blocks or unblocks a user.
func (i *Instance) UpdateBlocklist(ctx context.Context, jidStr, action string) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}

	var blockAction waEvents.BlocklistChangeAction
	switch action {
	case "block":
		blockAction = waEvents.BlocklistChangeActionBlock
	case "unblock":
		blockAction = waEvents.BlocklistChangeActionUnblock
	default:
		return fmt.Errorf("invalid action: %s (must be 'block' or 'unblock')", action)
	}

	_, err = i.client.UpdateBlocklist(ctx, jid, blockAction)
	return err
}

// GetProfile fetches a user's profile information.
func (i *Instance) GetProfile(ctx context.Context, jidStr string) (*ProfileInfo, error) {
	jid, err := parseJID(jidStr)
	if err != nil {
		return nil, err
	}

	profile := &ProfileInfo{JID: jid.String()}

	// Get profile picture
	pic, err := i.client.GetProfilePictureInfo(ctx, jid, &whatsmeow.GetProfilePictureParams{Preview: false})
	if err == nil && pic != nil {
		profile.PictureURL = pic.URL
	}

	return profile, nil
}

// --- Group operations ---

// CreateGroup creates a new WhatsApp group.
func (i *Instance) CreateGroup(ctx context.Context, name string, participants []string) (*CreateGroupResult, error) {
	jids := make([]waTypes.JID, len(participants))
	for idx, p := range participants {
		j, err := parseJID(p)
		if err != nil {
			return nil, fmt.Errorf("invalid participant %q: %w", p, err)
		}
		jids[idx] = j
	}

	info, err := i.client.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: jids,
	})
	if err != nil {
		return nil, fmt.Errorf("creating group: %w", err)
	}

	return &CreateGroupResult{
		JID:  info.JID.String(),
		Name: info.GroupName.Name,
	}, nil
}

type CreateGroupResult struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

// GetJoinedGroups returns all groups the instance is a member of.
func (i *Instance) GetJoinedGroups(ctx context.Context) ([]GroupInfoResult, error) {
	groups, err := i.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting joined groups: %w", err)
	}

	results := make([]GroupInfoResult, len(groups))
	for idx, g := range groups {
		results[idx] = groupInfoToResult(g)
	}
	return results, nil
}

type GroupInfoResult struct {
	JID          string                  `json:"jid"`
	Name         string                  `json:"name"`
	Topic        string                  `json:"topic"`
	OwnerJID     string                  `json:"owner_jid"`
	Participants []GroupParticipantResult `json:"participants"`
	Locked       bool                    `json:"locked"`
	Announce     bool                    `json:"announce"`
}

type GroupParticipantResult struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"is_admin"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

func groupInfoToResult(g *waTypes.GroupInfo) GroupInfoResult {
	participants := make([]GroupParticipantResult, len(g.Participants))
	for i, p := range g.Participants {
		participants[i] = GroupParticipantResult{
			JID:          p.JID.String(),
			IsAdmin:      p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		}
	}
	return GroupInfoResult{
		JID:          g.JID.String(),
		Name:         g.GroupName.Name,
		Topic:        g.Topic,
		OwnerJID:     g.OwnerJID.String(),
		Participants: participants,
		Locked:       g.IsLocked,
		Announce:     g.IsAnnounce,
	}
}

// GetGroupInfo gets info about a specific group.
func (i *Instance) GetGroupInfo(ctx context.Context, jidStr string) (*GroupInfoResult, error) {
	jid, err := parseJID(jidStr)
	if err != nil {
		return nil, err
	}

	info, err := i.client.GetGroupInfo(ctx, jid)
	if err != nil {
		return nil, fmt.Errorf("getting group info: %w", err)
	}

	result := groupInfoToResult(info)
	return &result, nil
}

// SetGroupName updates a group's name.
func (i *Instance) SetGroupName(ctx context.Context, jidStr, name string) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}
	return i.client.SetGroupName(ctx, jid, name)
}

// SetGroupDescription updates a group's topic/description.
func (i *Instance) SetGroupDescription(ctx context.Context, jidStr, desc string) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}
	return i.client.SetGroupTopic(ctx, jid, "", "", desc)
}

// SetGroupPhoto updates a group's photo.
func (i *Instance) SetGroupPhoto(ctx context.Context, jidStr string, photo []byte) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}
	_, err = i.client.SetGroupPhoto(ctx, jid, photo)
	return err
}

// UpdateGroupSettings updates group settings (locked, announce).
func (i *Instance) UpdateGroupSettings(ctx context.Context, jidStr string, locked, announce *bool) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}
	if locked != nil {
		if err := i.client.SetGroupLocked(ctx, jid, *locked); err != nil {
			return err
		}
	}
	if announce != nil {
		if err := i.client.SetGroupAnnounce(ctx, jid, *announce); err != nil {
			return err
		}
	}
	return nil
}

// ManageParticipants adds/removes/promotes/demotes participants.
func (i *Instance) ManageParticipants(ctx context.Context, groupJIDStr, action string, participants []string) error {
	groupJID, err := parseJID(groupJIDStr)
	if err != nil {
		return err
	}

	jids := make([]waTypes.JID, len(participants))
	for idx, p := range participants {
		j, err := parseJID(p)
		if err != nil {
			return fmt.Errorf("invalid participant %q: %w", p, err)
		}
		jids[idx] = j
	}

	var change whatsmeow.ParticipantChange
	switch action {
	case "add":
		change = whatsmeow.ParticipantChangeAdd
	case "remove":
		change = whatsmeow.ParticipantChangeRemove
	case "promote":
		change = whatsmeow.ParticipantChangePromote
	case "demote":
		change = whatsmeow.ParticipantChangeDemote
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	_, err = i.client.UpdateGroupParticipants(ctx, groupJID, jids, change)
	return err
}

// GetGroupInviteLink gets the invite link for a group.
func (i *Instance) GetGroupInviteLink(ctx context.Context, jidStr string) (string, error) {
	jid, err := parseJID(jidStr)
	if err != nil {
		return "", err
	}

	link, err := i.client.GetGroupInviteLink(ctx, jid, false)
	if err != nil {
		return "", fmt.Errorf("getting invite link: %w", err)
	}
	return link, nil
}

// LeaveGroup leaves a group.
func (i *Instance) LeaveGroup(ctx context.Context, jidStr string) error {
	jid, err := parseJID(jidStr)
	if err != nil {
		return err
	}
	return i.client.LeaveGroup(ctx, jid)
}

// --- Helpers ---

func parseJID(s string) (waTypes.JID, error) {
	if !strings.Contains(s, "@") {
		// Assume phone number, add default suffix
		s = s + "@s.whatsapp.net"
	}
	return waTypes.ParseJID(s)
}

const maxMediaSize = 50 << 20 // 50MB

var mediaClient = &http.Client{Timeout: 60 * time.Second}

func resolveMediaData(media string) ([]byte, error) {
	if strings.HasPrefix(media, "http://") || strings.HasPrefix(media, "https://") {
		resp, err := mediaClient.Get(media)
		if err != nil {
			return nil, fmt.Errorf("downloading media: %w", err)
		}
		defer resp.Body.Close()
		return io.ReadAll(io.LimitReader(resp.Body, maxMediaSize))
	}

	// Try base64 decode
	data, err := base64.StdEncoding.DecodeString(media)
	if err == nil && len(data) > 0 {
		return data, nil
	}

	return nil, fmt.Errorf("media must be a URL or base64 encoded data")
}

func buildMediaMessage(mediaType string, upload whatsmeow.UploadResponse, caption, filename, mimeType string, fileLen uint64) *waE2E.Message {
	switch mediaType {
	case "image":
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				Caption:       proto.String(caption),
				Mimetype:      proto.String(mimeType),
				URL:           proto.String(upload.URL),
				DirectPath:    proto.String(upload.DirectPath),
				MediaKey:      upload.MediaKey,
				FileEncSHA256: upload.FileEncSHA256,
				FileSHA256:    upload.FileSHA256,
				FileLength:    &fileLen,
			},
		}
	case "video":
		if mimeType == "" {
			mimeType = "video/mp4"
		}
		return &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				Caption:       proto.String(caption),
				Mimetype:      proto.String(mimeType),
				URL:           proto.String(upload.URL),
				DirectPath:    proto.String(upload.DirectPath),
				MediaKey:      upload.MediaKey,
				FileEncSHA256: upload.FileEncSHA256,
				FileSHA256:    upload.FileSHA256,
				FileLength:    &fileLen,
			},
		}
	case "audio":
		if mimeType == "" {
			mimeType = "audio/ogg; codecs=opus"
		}
		return &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				Mimetype:      proto.String(mimeType),
				URL:           proto.String(upload.URL),
				DirectPath:    proto.String(upload.DirectPath),
				MediaKey:      upload.MediaKey,
				FileEncSHA256: upload.FileEncSHA256,
				FileSHA256:    upload.FileSHA256,
				FileLength:    &fileLen,
				PTT:           proto.Bool(true),
			},
		}
	case "document":
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		return &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				Caption:       proto.String(caption),
				FileName:      proto.String(filename),
				Mimetype:      proto.String(mimeType),
				URL:           proto.String(upload.URL),
				DirectPath:    proto.String(upload.DirectPath),
				MediaKey:      upload.MediaKey,
				FileEncSHA256: upload.FileEncSHA256,
				FileSHA256:    upload.FileSHA256,
				FileLength:    &fileLen,
			},
		}
	}
	return &waE2E.Message{Conversation: proto.String("[unsupported media]")}
}

func buildVCard(name string, phones []string) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCARD\nVERSION:3.0\n")
	b.WriteString(fmt.Sprintf("FN:%s\n", name))
	for _, phone := range phones {
		b.WriteString(fmt.Sprintf("TEL;type=CELL;type=VOICE;waid=%s:%s\n", strings.TrimPrefix(phone, "+"), phone))
	}
	b.WriteString("END:VCARD")
	return b.String()
}

