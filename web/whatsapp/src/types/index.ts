// User types
export interface User {
  id: string;
  name: string;
  username: string;
  email?: string;
  avatar?: string;
  status?: string;
  last_seen?: string;
  is_group?: boolean;
  last_message?: string;
  last_message_time?: string;
  created_at?: string;
}

export interface Group {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
  member_ids: string[];
  created_at: string;
  avatar_url?: string;
  is_group: true;
}

export type Contact = User | Group;

// Message types
export interface Message {
  id: string;
  sender_id: string;
  sender_username?: string;
  receiver_id: string;
  group_id?: string;
  content: string;
  media_url?: string;
  created_at: string;
  status: 'sent' | 'delivered' | 'read';
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
  sender_username?: string;
  receiver_id?: string;
  content?: string;
  group_id?: string;
  media_url?: string;
  created_at?: string;
  status?: MessageStatus;
  type?: string;
  [key: string]: unknown;
}

// Typing indicator types
export interface TypingEvent {
  type: "typing";
  sender_id: string;
  receiver_id: string;
  is_typing: boolean;
  timestamp: string;
}
