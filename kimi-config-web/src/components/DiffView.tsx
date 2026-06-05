import { DiffEditor } from '@monaco-editor/react';

interface DiffViewProps {
  original: string;
  modified: string;
  originalLabel: string;
  modifiedLabel: string;
  emptyMessage?: string;
}

export default function DiffView({
  original,
  modified,
  originalLabel,
  modifiedLabel,
  emptyMessage,
}: DiffViewProps) {
  if (emptyMessage) {
    return (
      <div style={styles.empty}>
        {emptyMessage}
      </div>
    );
  }

  return (
    <div style={styles.panel}>
      <div style={styles.header}>
        <span>{originalLabel}</span>
        <span style={styles.arrow}>vs</span>
        <span>{modifiedLabel}</span>
      </div>
      <DiffEditor
        height="100%"
        language="python"
        original={original}
        modified={modified}
        theme="vs-dark"
        options={{
          readOnly: true,
          renderSideBySide: false,
          minimap: { enabled: false },
          fontSize: 12,
          fontFamily: 'JetBrains Mono, Fira Code, monospace',
          scrollBeyondLastLine: false,
          automaticLayout: true,
          padding: { top: 12 },
        }}
      />
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
    height: 36,
    padding: '0 12px',
    borderBottom: '1px solid #333',
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    color: '#ccc',
    fontSize: 12,
    whiteSpace: 'nowrap',
  },
  arrow: {
    color: '#888',
  },
  empty: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
    color: '#888',
    fontSize: 13,
    textAlign: 'center',
  },
};
