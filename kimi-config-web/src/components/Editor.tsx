import type { PlatformVersion } from '../api';
import Editor from '@monaco-editor/react';

interface EditorProps {
  value: string;
  onChange: (value: string) => void;
  label?: string;
  versions?: PlatformVersion[];
  currentVersion?: string;
  onVersionSelect?: (version: string) => void;
  readOnly?: boolean;
}

export default function ScriptEditor({
  value,
  onChange,
  label = 'script.star',
  versions = [],
  currentVersion = '',
  onVersionSelect,
  readOnly = false,
}: EditorProps) {
  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
      <div style={styles.tabBar}>
        <div style={styles.tabGroup}>
          <span style={styles.tab}>{label}</span>
          {readOnly && <span style={styles.readOnlyBadge}>Read-only</span>}
        </div>
        {versions.length > 0 && (
          <label style={styles.versionPicker}>
            Version
            <select
              value={currentVersion}
              onChange={(event) => onVersionSelect?.(event.target.value)}
              style={styles.select}
            >
              {versions.map((item) => (
                <option key={`${item.path}-${item.version}`} value={item.version}>
                  {item.version}{item.latest ? ' latest' : ''}{item.draft ? ' draft' : ''}
                </option>
              ))}
              {currentVersion && !versions.some((item) => item.version === currentVersion) && (
                <option value={currentVersion}>{currentVersion}</option>
              )}
            </select>
          </label>
        )}
      </div>
      <Editor
        height="100%"
        defaultLanguage="python"
        value={value}
        onChange={(v) => {
          if (!readOnly) onChange(v || '');
        }}
        theme="vs-dark"
        options={{
          readOnly,
          domReadOnly: readOnly,
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
    justifyContent: 'space-between',
    padding: '0 12px',
  },
  tab: {
    padding: '6px 16px',
    backgroundColor: '#1e1e1e',
    color: '#fff',
    fontSize: 12,
    borderTop: '2px solid #007acc',
  },
  tabGroup: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    minWidth: 0,
  },
  readOnlyBadge: {
    padding: '2px 6px',
    backgroundColor: '#3c3c3c',
    border: '1px solid #555',
    borderRadius: 3,
    color: '#bbb',
    fontSize: 11,
  },
  versionPicker: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    color: '#888',
    fontSize: 11,
    textTransform: 'uppercase',
  },
  select: {
    height: 24,
    backgroundColor: '#252526',
    color: '#fff',
    border: '1px solid #444',
    borderRadius: 3,
    fontSize: 12,
    textTransform: 'none',
  },
};
