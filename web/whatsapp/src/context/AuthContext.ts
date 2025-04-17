import { createContext } from "react";
import { User, AuthState, LoginCredentials, RegisterData } from "@/types";

export interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  updateUserProfile: (userData: Partial<User>) => void;
  authLoaded: boolean;
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  token: null,
  isAuthenticated: false,
  authLoaded: false,
  login: async () => {},
  register: async () => {},
  logout: () => {},
  updateUserProfile: () => {},
});
