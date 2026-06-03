interface SidebarProps {
  platforms: string[];
  current: string;
  onSelect: (platform: string) => void;
}

export default function Sidebar({ platforms, current, onSelect }: SidebarProps) {
  return (
    <div style={styles.sidebar}>
      <div style={styles.header}>Platforms</div>
      <div style={styles.list}>
        {platforms.map((p) => (
          <button
            key={p}
            onClick={() => onSelect(p)}
            style={{
              ...styles.item,
              ...(p === current ? styles.itemActive : {}),
            }}
          >
            {p}
          </button>
        ))}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  sidebar: {
    width: 140,
    minWidth: 140,
    backgroundColor: '#1e1e1e',
    borderRight: '1px solid #333',
    display: 'flex',
    flexDirection: 'column',
  },
  header: {
    padding: '10px 12px',
    fontSize: 11,
    fontWeight: 600,
    color: '#888',
    textTransform: 'uppercase',
    letterSpacing: 1,
    borderBottom: '1px solid #333',
  },
  list: {
    padding: 6,
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
  },
  item: {
    padding: '6px 10px',
    borderRadius: 4,
    border: 'none',
    background: 'transparent',
    color: '#ccc',
    fontSize: 13,
    fontFamily: 'monospace',
    textAlign: 'left',
    cursor: 'pointer',
  },
  itemActive: {
    backgroundColor: '#37373d',
    color: '#fff',
  },
};
