import { useState } from 'react';
import { api } from '../api';
import type { ScriptContext } from '../api';

interface PreviewProps {
  script: string;
}

export default function Preview({ script }: PreviewProps) {
  const [ctx, setCtx] = useState<ScriptContext>({
    platform: 'ios',
    version: '2.5.5',
    language: 'zh',
    region: 'domestic',
  });
  const [result, setResult] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const runPreview = async () => {
    setLoading(true);
    setError('');
    try {
      const resp = await api.preview(script, ctx);
      if (resp.data.error) {
        setError(resp.data.error);
        setResult('');
      } else {
        setResult(JSON.stringify(resp.data.config, null, 2));
      }
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } };
      setError(err.response?.data?.error || String(e));
      setResult('');
    } finally {
      setLoading(false);
    }
  };

  const update = (field: keyof ScriptContext, value: string) => {
    setCtx((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div style={styles.panel}>
      <div style={styles.controls}>
        {(['platform', 'version', 'language', 'region'] as const).map((field) => (
          <div key={field} style={styles.field}>
            <label style={styles.label}>{field}</label>
            <input
              style={styles.input}
              value={ctx[field]}
              onChange={(e) => update(field, e.target.value)}
            />
          </div>
        ))}
        <button onClick={runPreview} disabled={loading} style={styles.button}>
          {loading ? 'Running...' : 'Run Preview'}
        </button>
      </div>

      {error && (
        <pre style={styles.error}>{error}</pre>
      )}

      {result && (
        <pre style={styles.result}>{result}</pre>
      )}
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
  controls: {
    padding: 16,
    display: 'flex',
    flexDirection: 'column',
    gap: 10,
    borderBottom: '1px solid #333',
  },
  field: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
  },
  label: {
    width: 60,
    fontSize: 11,
    color: '#888',
    textTransform: 'uppercase',
  },
  input: {
    flex: 1,
    padding: '5px 8px',
    backgroundColor: '#3c3c3c',
    border: '1px solid #555',
    borderRadius: 3,
    color: '#fff',
    fontSize: 13,
    fontFamily: 'monospace',
  },
  button: {
    padding: '8px 16px',
    backgroundColor: '#0e639c',
    color: '#fff',
    border: 'none',
    borderRadius: 3,
    fontSize: 13,
    cursor: 'pointer',
    marginTop: 4,
  },
  error: {
    padding: 16,
    color: '#f48771',
    fontSize: 12,
    fontFamily: 'monospace',
    whiteSpace: 'pre-wrap',
    margin: 0,
    overflow: 'auto',
  },
  result: {
    flex: 1,
    padding: 16,
    color: '#ce9178',
    fontSize: 12,
    fontFamily: 'monospace',
    whiteSpace: 'pre-wrap',
    margin: 0,
    overflow: 'auto',
    backgroundColor: '#1e1e1e',
  },
};
