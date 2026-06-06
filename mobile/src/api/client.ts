import AsyncStorage from '@react-native-async-storage/async-storage';

const BASE_URL = 'http://192.168.0.108:8080';

const STORAGE_KEYS = {
  token: 'trace_jwt',
  userId: 'trace_user_id',
  interviewDone: 'trace_interview_done',
};

export const Storage = {
  async getToken(): Promise<string | null> {
    return AsyncStorage.getItem(STORAGE_KEYS.token);
  },
  async setToken(token: string): Promise<void> {
    await AsyncStorage.setItem(STORAGE_KEYS.token, token);
  },
  async getUserId(): Promise<string | null> {
    return AsyncStorage.getItem(STORAGE_KEYS.userId);
  },
  async setUserId(id: string): Promise<void> {
    await AsyncStorage.setItem(STORAGE_KEYS.userId, id);
  },
  async isInterviewDone(): Promise<boolean> {
    const v = await AsyncStorage.getItem(STORAGE_KEYS.interviewDone);
    return v === 'true';
  },
  async setInterviewDone(): Promise<void> {
    await AsyncStorage.setItem(STORAGE_KEYS.interviewDone, 'true');
  },
  async clear(): Promise<void> {
    await AsyncStorage.multiRemove(Object.values(STORAGE_KEYS));
  },
};

async function authFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const token = await Storage.getToken();
  return fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers as Record<string, string> || {}),
    },
  });
}

export const API = {
  async login(userId: string): Promise<string> {
    const res = await fetch(`${BASE_URL}/api/v1/auth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: userId }),
    });
    if (!res.ok) throw new Error('Login failed');
    const data = await res.json();
    await Storage.setToken(data.token);
    await Storage.setUserId(userId);
    return data.token;
  },

  async startInterview(): Promise<{ session_id: string; opening_message: string }> {
    const res = await authFetch('/api/v1/interview/start', { method: 'POST' });
    if (!res.ok) throw new Error('Failed to start interview');
    return res.json();
  },

  async sendTurn(sessionId: string, message: string): Promise<{
    agent_reply: string;
    is_complete: boolean;
    progress_pct: number;
    extracted_fields: Array<{ field_name: string; confidence: number; filled: boolean }>;
  }> {
    const res = await authFetch('/api/v1/interview/turn', {
      method: 'POST',
      body: JSON.stringify({ session_id: sessionId, user_message: message }),
    });
    if (!res.ok) throw new Error('Turn failed');
    return res.json();
  },

  async confirmInterview(sessionId: string): Promise<{ profile_summary: string; graph_build_started: boolean }> {
    const res = await authFetch('/api/v1/interview/confirm', {
      method: 'POST',
      body: JSON.stringify({ session_id: sessionId, confirmed: true }),
    });
    if (!res.ok) throw new Error('Confirm failed');
    return res.json();
  },

  async getGraph(): Promise<{
    user_id: string;
    version: number;
    nodes: Array<{ id: string; type: string; label: string; x: number; y: number }>;
    edges: Array<{ id: string; from_node: string; to_node: string; relation: string; polarity: string; confidence: number }>;
    last_updated: number;
  }> {
    const res = await authFetch('/api/v1/graph');
    if (!res.ok) throw new Error('Failed to fetch graph');
    return res.json();
  },

  async getWorkoutToday(): Promise<any> {
    const res = await authFetch('/api/v1/workout/today');
    if (!res.ok) throw new Error('No workout yet');
    return res.json();
  },
};
