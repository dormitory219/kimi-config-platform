import axios from 'axios';

export interface ScriptContext {
  platform: string;
  version: string;
  language: string;
  region: string;
}

export interface PlatformVersion {
  version: string;
  path: string;
  latest: boolean;
  legacy: boolean;
  draft: boolean;
}

export interface VersionsResponse {
  platform: string;
  versions: PlatformVersion[];
  latest?: PlatformVersion;
  nextVersion: string;
}

export interface DraftResponse {
  platform: string;
  version: string;
  path: string;
  content: string;
  baseVersion: string;
  basePath: string;
  baseContent: string;
  latestVersion: string;
  workingVersion: string;
}

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const api = {
  getPlatforms: () => apiClient.get<{ platforms: string[] }>('/api/platforms'),

  getScript: (platform: string) =>
    apiClient.get<{ platform: string; content: string }>(`/api/scripts/${platform}`),

  getVersionedScript: (platform: string, version: string) =>
    apiClient.get<{ platform: string; version: string; path: string; content: string }>(
      `/api/scripts/${platform}`,
      { params: { version } },
    ),

  getVersions: (platform: string) =>
    apiClient.get<VersionsResponse>(`/api/scripts/${platform}/versions`),

  createDraft: (platform: string) =>
    apiClient.post<DraftResponse>(`/api/scripts/${platform}/drafts`),

  saveScript: (platform: string, content: string, version?: string) =>
    apiClient.post(`/api/scripts/${platform}`, { content }, { params: { version } }),

  publishScript: (platform: string, message?: string, version?: string) =>
    apiClient.post(`/api/scripts/${platform}/publish`, { message }, { params: { version } }),

  getHistory: (platform: string, version?: string) =>
    apiClient.get<{ platform: string; commits: { hash: string; message: string; author: string; timestamp: string }[] }>(
      `/api/scripts/${platform}/history`,
      { params: { version } },
    ),

  preview: (script: string, ctx: ScriptContext) =>
    apiClient.post<{ config: Record<string, unknown>; error?: string }>('/api/preview', {
      script,
      ctx,
    }),
};
