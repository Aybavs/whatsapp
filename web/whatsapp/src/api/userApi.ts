import axiosInstance from '@/api/axios';
import { User } from '@/types';

export const userApi = {
  searchUsers: async (query?: string): Promise<User[]> => {
    const response = await axiosInstance.get<User[]>("/users/search", {
      params: query ? { query: query } : undefined,
    });
    return response.data;
  },

  getUserById: async (id: string): Promise<User> => {
    const response = await axiosInstance.get<User>(`/users/${id}`);
    return response.data;
  },

  updateProfile: async (id: string, data: Partial<User>): Promise<User> => {
    const response = await axiosInstance.put<User>(`/users/${id}`, data);
    return response.data;
  },

  updateStatus: async (id: string, status: string): Promise<User> => {
    const response = await axiosInstance.patch<User>(`/users/${id}/status`, {
      status,
    });
    return response.data;
  },
  getContacts: async (): Promise<User[]> => {
    const response = await axiosInstance.get<User[]>("/users/contacts");
    return response.data;
  },

  addContact: async (contactId: string): Promise<void> => {
    await axiosInstance.post("/users/contacts", { contact_id: contactId });
  },
};
