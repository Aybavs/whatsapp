import { AuthResponse, LoginCredentials, RegisterData, User } from '@/types';

import axiosInstance from './axios';

export const authApi = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>(
      "/users/login",
      credentials
    );
    return response.data;
  },

  register: async (userData: RegisterData): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>(
      "/users/register",
      userData
    );
    return response.data;
  },

  getProfile: async (UserID: string): Promise<User> => {
    const response = await axiosInstance.get<User>(`/users/${UserID}`);
    return response.data;
  },

  // Kimliği doğrulanmış kullanıcı için profil bilgisi getirme
  // (Opsiyonel: token ile doğrulanmış kullanıcının kendi profilini almak istiyorsanız)
  getCurrentUserProfile: async (): Promise<User> => {
    const response = await axiosInstance.get<User>("/users/me");
    return response.data;
  },
};
