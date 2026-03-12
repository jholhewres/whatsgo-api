import QRCode from 'react-qr-code';

interface Props {
  qrCode: string | null;
  pairingCode: string | null;
  onClose: () => void;
}

export default function QRCodeModal({ qrCode, pairingCode, onClose }: Props) {
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Connect Instance</h3>
          <button className="btn-close" onClick={onClose}>&times;</button>
        </div>
        <div className="modal-body">
          {qrCode && (
            <div className="qr-container">
              <p>Scan this QR code with WhatsApp</p>
              <div className="qr-code">
                <QRCode value={qrCode} size={256} />
              </div>
            </div>
          )}
          {pairingCode && (
            <div className="pairing-container">
              <p>Enter this code on your phone</p>
              <div className="pairing-code">{pairingCode}</div>
            </div>
          )}
          {!qrCode && !pairingCode && (
            <div className="loading-container">
              <div className="spinner" />
              <p>Waiting for QR code...</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
