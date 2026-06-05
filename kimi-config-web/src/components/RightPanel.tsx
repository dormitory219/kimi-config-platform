import { useState } from 'react';
import Preview from './Preview';
import History from './History';
import DiffView from './DiffView';

interface RightPanelProps {
  script: string;
  platform: string;
  version: string;
  diffBaseVersion: string;
  diffBaseContent: string;
}

type Tab = 'preview' | 'diff' | 'history';

export default function RightPanel({
  script,
  platform,
  version,
  diffBaseVersion,
  diffBaseContent,
}: RightPanelProps) {
  const [activeTab, setActiveTab] = useState<Tab>('preview');
  const hasDiffBase = Boolean(diffBaseVersion);

  return (
    <div style={styles.panel}>
      <div style={styles.tabs}>
        <button
          onClick={() => setActiveTab('preview')}
          style={{
            ...styles.tab,
            ...(activeTab === 'preview' ? styles.tabActive : {}),
          }}
        >
          Preview
        </button>
        <button
          onClick={() => setActiveTab('history')}
          style={{
            ...styles.tab,
            ...(activeTab === 'history' ? styles.tabActive : {}),
          }}
        >
          History
        </button>
        <button
          onClick={() => setActiveTab('diff')}
          style={{
            ...styles.tab,
            ...(activeTab === 'diff' ? styles.tabActive : {}),
          }}
        >
          Diff
        </button>
      </div>
      <div style={styles.content}>
        {activeTab === 'preview' && (
          <Preview script={script} platform={platform} />
        )}
        {activeTab === 'history' && <History platform={platform} version={version} />}
        {activeTab === 'diff' && (
          <DiffView
            original={diffBaseContent}
            modified={script}
            originalLabel={`base ${diffBaseVersion || '-'}`}
            modifiedLabel={`editing ${version || '-'}`}
            emptyMessage={!hasDiffBase ? `${version || 'This version'} is the first version and has no diff.` : undefined}
          />
        )}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    width: 380,
    minWidth: 380,
    backgroundColor: '#1e1e1e',
    borderLeft: '1px solid #333',
    display: 'flex',
    flexDirection: 'column',
  },
  tabs: {
    display: 'flex',
    borderBottom: '1px solid #333',
    height: 36,
  },
  tab: {
    flex: 1,
    padding: '8px 0',
    backgroundColor: 'transparent',
    border: 'none',
    color: '#888',
    fontSize: 12,
    fontWeight: 600,
    textTransform: 'uppercase',
    letterSpacing: 1,
    cursor: 'pointer',
  },
  tabActive: {
    color: '#fff',
    borderBottom: '2px solid #0e639c',
    backgroundColor: '#252526',
  },
  content: {
    flex: 1,
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column',
  },
};
