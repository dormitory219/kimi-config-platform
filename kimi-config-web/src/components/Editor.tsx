import Editor from '@monaco-editor/react';

interface EditorProps {
  value: string;
  onChange: (value: string) => void;
}

export default function ScriptEditor({ value, onChange }: EditorProps) {
  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
      <div style={styles.tabBar}>
        <span style={styles.tab}>script.star</span>
      </div>
      <Editor
        height="100%"
        defaultLanguage="python"
        value={value}
        onChange={(v) => onChange(v || '')}
        theme="vs-dark"
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          fontFamily: 'JetBrains Mono, Fira Code, monospace',
          lineNumbers: 'on',
          roundedSelection: false,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 4,
          insertSpaces: true,
          padding: { top: 16 },
        }}
      />
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  tabBar: {
    height: 36,
    backgroundColor: '#2d2d30',
    borderBottom: '1px solid #333',
    display: 'flex',
    alignItems: 'center',
    paddingLeft: 12,
  },
  tab: {
    padding: '6px 16px',
    backgroundColor: '#1e1e1e',
    color: '#fff',
    fontSize: 12,
    borderTop: '2px solid #007acc',
  },
};
