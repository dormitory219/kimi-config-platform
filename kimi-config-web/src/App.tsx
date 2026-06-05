import { useEffect, useState, useCallback } from 'react';
import { api, type PlatformVersion } from './api';
import Sidebar from './components/Sidebar';
import ScriptEditor from './components/Editor';
import RightPanel from './components/RightPanel';

export default function App() {
  const [platforms, setPlatforms] = useState<string[]>([]);
  const [currentPlatform, setCurrentPlatform] = useState('');
  const [versions, setVersions] = useState<PlatformVersion[]>([]);
  const [currentVersion, setCurrentVersion] = useState('');
  const [diffBaseVersion, setDiffBaseVersion] = useState('');
  const [diffBaseContent, setDiffBaseContent] = useState('');
  const [scriptContent, setScriptContent] = useState('');
  const [saved, setSaved] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [drafting, setDrafting] = useState(false);
  const [message, setMessage] = useState('');
  const currentVersionMeta = versions.find((item) => item.version === currentVersion);
  const isCurrentDraft = Boolean(currentVersionMeta?.draft);
  const isViewingOldVersion = Boolean(currentVersion && currentVersionMeta && !currentVersionMeta.latest && !currentVersionMeta.draft);
  const canCreateDraft = Boolean(currentPlatform && currentVersionMeta?.latest && !drafting);
  const canSave = Boolean(currentPlatform && isCurrentDraft);
  const canPublish = Boolean(currentPlatform && (currentVersionMeta?.latest || isCurrentDraft) && !publishing);

  // Load platforms
  useEffect(() => {
    api.getPlatforms().then((resp) => {
      const list = resp.data.platforms;
      setPlatforms(list);
      if (list.length > 0 && !currentPlatform) {
        setCurrentPlatform(list.includes('ios') ? 'ios' : list[0]);
      }
    });
  }, []);

  const loadPlatformWorkspace = useCallback(async (platform: string) => {
    const versionsResp = await api.getVersions(platform);
    const nextVersions = versionsResp.data.versions || [];
    const latest = versionsResp.data.latest || nextVersions[nextVersions.length - 1];

    setVersions(nextVersions);

    if (!latest) {
      setCurrentVersion('');
      setDiffBaseVersion('');
      setDiffBaseContent('');
      setScriptContent('');
      setSaved(true);
      return;
    }

    const scriptResp = await api.getVersionedScript(platform, latest.version);
    const baseVersion = findPreviousVersion(latest.version, nextVersions);
    if (baseVersion) {
      const baseResp = await api.getVersionedScript(platform, baseVersion);
      setDiffBaseVersion(baseVersion);
      setDiffBaseContent(baseResp.data.content);
    } else {
      setDiffBaseVersion('');
      setDiffBaseContent('');
    }
    setCurrentVersion(latest.version);
    setScriptContent(scriptResp.data.content);
    setSaved(true);
  }, []);

  // Load latest version when platform changes
  useEffect(() => {
    if (!currentPlatform) return;
    loadPlatformWorkspace(currentPlatform).catch((err) => {
      setMessage(err.response?.data?.error || 'Failed to load platform');
      setTimeout(() => setMessage(''), 2200);
    });
  }, [currentPlatform, loadPlatformWorkspace]);

  const handleSave = useCallback(async () => {
    if (!currentPlatform) return;
    if (!canSave) {
      setMessage(isViewingOldVersion ? 'Older versions are read-only' : 'Create a draft before saving changes');
      setTimeout(() => setMessage(''), 2200);
      return;
    }
    await api.saveScript(currentPlatform, scriptContent, currentVersion);
    await loadPlatformWorkspace(currentPlatform);
    setSaved(true);
    setMessage('Saved');
    setTimeout(() => setMessage(''), 1500);
  }, [canSave, currentPlatform, currentVersion, isViewingOldVersion, loadPlatformWorkspace, scriptContent]);

  const handleCreateDraft = useCallback(async () => {
    if (!currentPlatform) return;
    if (!canCreateDraft) {
      setMessage('Select the latest version before creating a draft');
      setTimeout(() => setMessage(''), 2200);
      return;
    }
    setDrafting(true);
    try {
      const resp = await api.createDraft(currentPlatform);
      setCurrentVersion(resp.data.version);
      setDiffBaseVersion(resp.data.baseVersion);
      setDiffBaseContent(resp.data.baseContent);
      setScriptContent(resp.data.content);
      setSaved(true);
      setVersions((prev) => [
        ...prev.filter((item) => item.version !== resp.data.version),
        {
          version: resp.data.version,
          path: resp.data.path,
          latest: false,
          legacy: false,
          draft: true,
        },
      ]);
      setMessage(`Draft ${resp.data.version} created from ${resp.data.baseVersion}`);
      setTimeout(() => setMessage(''), 2200);
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } };
      setMessage(err.response?.data?.error || 'Create draft failed');
    } finally {
      setDrafting(false);
    }
  }, [canCreateDraft, currentPlatform]);

  const handleVersionSelect = useCallback(async (version: string) => {
    if (!currentPlatform || !version) return;
    const resp = await api.getVersionedScript(currentPlatform, version);
    const baseVersion = findPreviousVersion(version, versions);
    if (baseVersion) {
      const baseResp = await api.getVersionedScript(currentPlatform, baseVersion);
      setDiffBaseVersion(baseVersion);
      setDiffBaseContent(baseResp.data.content);
    } else {
      setDiffBaseVersion('');
      setDiffBaseContent('');
    }
    setCurrentVersion(version);
    setScriptContent(resp.data.content);
    setSaved(true);
  }, [currentPlatform, versions]);

  const handlePublish = useCallback(async () => {
    if (!currentPlatform) return;
    if (!canPublish) {
      setMessage('Older versions are read-only');
      setTimeout(() => setMessage(''), 2200);
      return;
    }
    setPublishing(true);
    try {
      if (!saved) {
        await api.saveScript(currentPlatform, scriptContent, currentVersion);
      }
      await api.publishScript(currentPlatform, undefined, currentVersion);
      await loadPlatformWorkspace(currentPlatform);
      setMessage('Published!');
      setTimeout(() => setMessage(''), 2000);
    } catch (e) {
      setMessage('Publish failed');
    } finally {
      setPublishing(false);
    }
  }, [canPublish, currentPlatform, currentVersion, loadPlatformWorkspace, saved, scriptContent]);

  return (
    <div style={styles.app}>
      {/* Header */}
      <div style={styles.header}>
        <div style={styles.title}>Kimi Config Platform</div>
        <div style={styles.actions}>
          {message && <span style={styles.message}>{message}</span>}
          {!saved && <span style={styles.unsaved}>unsaved</span>}
          <button
            onClick={handleCreateDraft}
            disabled={!canCreateDraft}
            style={{
              ...styles.btnSecondary,
              ...(!canCreateDraft ? styles.btnDisabled : {}),
            }}
          >
            {drafting ? 'Creating...' : 'New Draft'}
          </button>
          <button
            onClick={handleSave}
            disabled={!canSave}
            style={{
              ...styles.btnSecondary,
              ...(!canSave ? styles.btnDisabled : {}),
            }}
          >
            Save
          </button>
          <button
            onClick={handlePublish}
            disabled={!canPublish}
            style={{
              ...styles.btnPrimary,
              ...(!canPublish ? styles.btnDisabled : {}),
            }}
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
          label={`${currentPlatform || 'script'}/${currentVersion || 'latest'}.star`}
          versions={versions}
          currentVersion={currentVersion}
          onVersionSelect={handleVersionSelect}
          readOnly={!isCurrentDraft}
          onChange={(v) => {
            if (!currentVersion || !isCurrentDraft) return;
            setScriptContent(v);
            setSaved(false);
          }}
        />
        <RightPanel
          script={scriptContent}
          platform={currentPlatform}
          version={currentVersion}
          diffBaseVersion={diffBaseVersion}
          diffBaseContent={diffBaseContent}
        />
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
  btnDisabled: {
    opacity: 0.45,
    cursor: 'not-allowed',
  },
  main: {
    flex: 1,
    display: 'flex',
    overflow: 'hidden',
  },
};

function findPreviousVersion(version: string, versions: PlatformVersion[]): string {
  const currentNumber = parseVersionNumber(version);
  if (currentNumber <= 1) return '';

  const candidates = versions
    .map((item) => item.version)
    .filter((item) => parseVersionNumber(item) < currentNumber)
    .sort((a, b) => parseVersionNumber(b) - parseVersionNumber(a));

  return candidates[0] || '';
}

function parseVersionNumber(version: string): number {
  const parsed = Number.parseInt(version.replace(/^v/, ''), 10);
  return Number.isFinite(parsed) ? parsed : 0;
}
