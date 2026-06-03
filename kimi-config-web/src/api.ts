import axios from 'axios';

export interface ScriptContext {
  platform: string;
  version: string;
  language: string;
  region: string;
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

  saveScript: (platform: string, content: string) =>
    apiClient.post(`/api/scripts/${platform}`, { content }),

  publishScript: (platform: string, message?: string) =>
    apiClient.post(`/api/scripts/${platform}/publish`, { message }),

  getHistory: (platform: string) =>
    apiClient.get<{ platform: string; commits: { hash: string; message: string; author: string; timestamp: string }[] }>(`/api/scripts/${platform}/history`),

  preview: (script: string, ctx: ScriptContext) =>
    apiClient.post<{ config: Record<string, unknown>; error?: string }>('/api/preview', {
      script,
      ctx,
    }),
};
