import { useEffect, useState, useCallback } from 'react';
import { api } from './api';
import Sidebar from './components/Sidebar';
import ScriptEditor from './components/Editor';
import RightPanel from './components/RightPanel';

export default function App() {
  const [platforms, setPlatforms] = useState<string[]>([]);
  const [currentPlatform, setCurrentPlatform] = useState('');
  const [scriptContent, setScriptContent] = useState('');
  const [saved, setSaved] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [message, setMessage] = useState('');

  // Load platforms
  useEffect(() => {
    api.getPlatforms().then((resp) => {
      const list = resp.data.platforms;
      setPlatforms(list);
      if (list.length > 0 && !currentPlatform) {
        setCurrentPlatform(list[0]);
      }
    });
  }, []);

  // Load script when platform changes
  useEffect(() => {
    if (!currentPlatform) return;
    api.getScript(currentPlatform).then((resp) => {
      setScriptContent(resp.data.content);
      setSaved(true);
    });
  }, [currentPlatform]);

  const handleSave = useCallback(async () => {
    if (!currentPlatform) return;
    await api.saveScript(currentPlatform, scriptContent);
    setSaved(true);
    setMessage('Saved');
    setTimeout(() => setMessage(''), 1500);
  }, [currentPlatform, scriptContent]);

  const handlePublish = useCallback(async () => {
    if (!currentPlatform) return;
    setPublishing(true);
    try {
      await api.publishScript(currentPlatform);
      setMessage('Published!');
      setTimeout(() => setMessage(''), 2000);
    } catch (e) {
      setMessage('Publish failed');
    } finally {
      setPublishing(false);
    }
  }, [currentPlatform]);

  return (
    <div style={styles.app}>
      {/* Header */}
      <div style={styles.header}>
        <div style={styles.title}>Kimi Config Platform</div>
        <div style={styles.actions}>
          {message && <span style={styles.message}>{message}</span>}
          {!saved && <span style={styles.unsaved}>unsaved</span>}
          <button onClick={handleSave} style={styles.btnSecondary}>
            Save
          </button>
          <button
            onClick={handlePublish}
            disabled={publishing}
            style={styles.btnPrimary}
          >
            {publishing ? 'Publishing...' : 'Publish'}
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div style={styles.main}>
        <Sidebar
          platforms={platforms}
          current={currentPlatform}
          onSelect={setCurrentPlatform}
        />
        <ScriptEditor
          value={scriptContent}
          onChange={(v) => {
            setScriptContent(v);
            setSaved(false);
          }}
        />
        <RightPanel script={scriptContent} platform={currentPlatform} />
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  app: {
    display: 'flex',
    flexDirection: 'column',
    height: '100vh',
    backgroundColor: '#1e1e1e',
    color: '#ccc',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
  },
  header: {
    height: 48,
    backgroundColor: '#2d2d30',
    borderBottom: '1px solid #333',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 16px',
  },
  title: {
    fontSize: 15,
    fontWeight: 600,
    color: '#fff',
  },
  actions: {
    display: 'flex',
    alignItems: 'center',
    gap: 12,
  },
  message: {
    fontSize: 12,
    color: '#89d185',
  },
  unsaved: {
    fontSize: 11,
    color: '#f48771',
    fontStyle: 'italic',
  },
  btnSecondary: {
    padding: '5px 14px',
    backgroundColor: '#3c3c3c',
    color: '#fff',
    border: '1px solid #555',
    borderRadius: 3,
    fontSize: 12,
    cursor: 'pointer',
  },
  btnPrimary: {
    padding: '5px 14px',
    backgroundColor: '#0e639c',
    color: '#fff',
    border: 'none',
    borderRadius: 3,
    fontSize: 12,
    cursor: 'pointer',
  },
  main: {
    flex: 1,
    display: 'flex',
    overflow: 'hidden',
  },
};