interface Props {
  status: string;
}

export default function StatusBadge({ status }: Props) {
  const colorMap: Record<string, string> = {
    open: '#22c55e',
    connecting: '#eab308',
    close: '#ef4444',
  };

  const color = colorMap[status] || '#6b7280';

  return (
    <span className="badge" style={{ backgroundColor: color }}>
      {status}
    </span>
  );
}
