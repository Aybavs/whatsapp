// User types
export interface User {
  id: string;
  name: string;
  username: string;
  email?: string;
  avatar?: string;
  status?: string;
  last_seen?: string;
}

// Message types
export interface Message {
  id: string;
  sender_id: string;
  receiver_id: string;
  content: string;
  media_url?: string;
  created_at: string;
  status: MessageStatus;
}

export type MessageStatus = "sent" | "delivered" | "read";

// Authentication types
export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
}

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface RegisterData {
  name: string;
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// WebSocket types
export interface WebSocketMessage {
  id?: string;
  sender_id?: string;
  receiver_id?: string;
  content?: string;
  media_url?: string;
  created_at?: string;
  status?: MessageStatus;
  type?: string;
  [key: string]: unknown;
}
