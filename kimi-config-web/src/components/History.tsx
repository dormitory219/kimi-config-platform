import { useEffect, useState } from 'react';
import { api } from '../api';

interface HistoryProps {
  platform: string;
}

interface Commit {
  hash: string;
  message: string;
  author: string;
  timestamp: string;
}

export default function History({ platform }: HistoryProps) {
  const [commits, setCommits] = useState<Commit[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!platform) return;
    setLoading(true);
    api.getHistory(platform)
      .then((resp) => setCommits(resp.data.commits))
      .catch(() => setCommits([]))
      .finally(() => setLoading(false));
  }, [platform]);

  const formatTime = (ts: string) => {
    const d = new Date(ts);
    return d.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div style={styles.panel}>
      <div style={styles.header}>History</div>
      {loading && <div style={styles.empty}>Loading...</div>}
      {!loading && commits.length === 0 && (
        <div style={styles.empty}>No history yet</div>
      )}
      <div style={styles.list}>
        {commits.map((c) => (
          <div key={c.hash} style={styles.commit}>
            <div style={styles.commitMessage}>{c.message}</div>
            <div style={styles.commitMeta}>
              <span style={styles.hash}>{c.hash}</span>
              <span>{c.author}</span>
              <span>{formatTime(c.timestamp)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden',
  },
  header: {
    padding: '12px 16px',
    fontSize: 12,
    fontWeight: 600,
    color: '#888',
    textTransform: 'uppercase',
    letterSpacing: 1,
    borderBottom: '1px solid #333',
  },
  list: {
    flex: 1,
    overflow: 'auto',
    padding: '8px 0',
  },
  commit: {
    padding: '8px 16px',
    borderBottom: '1px solid #2a2a2a',
  },
  commitMessage: {
    fontSize: 12,
    color: '#ccc',
    marginBottom: 4,
    wordBreak: 'break-word',
  },
  commitMeta: {
    fontSize: 11,
    color: '#666',
    display: 'flex',
    gap: 8,
    alignItems: 'center',
  },
  hash: {
    fontFamily: 'monospace',
    color: '#0e639c',
  },
  empty: {
    padding: 24,
    fontSize: 12,
    color: '#666',
    textAlign: 'center',
  },
};
