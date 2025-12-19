import axiosInstance from '@/api/axios';
import { Message } from '@/types';

export const messageApi = {
  getMessages: async (targetId: string, isGroup: boolean = false): Promise<Message[]> => {
    const url = `/messages/${targetId}`;
    const response = await axiosInstance.get<Message[]>(url);
    return response.data;
  },

  sendMessage: async (
    targetId: string,
    content: string,
    mediaUrl?: string,
    isGroup: boolean = false
  ): Promise<Message> => {
    const payload:any = {
      content,
      media_url: mediaUrl,
    };

    if (isGroup) {
      payload.group_id = targetId;
    } else {
      payload.receiver_id = targetId;
    }

    const response = await axiosInstance.post<Message>("/messages", payload);
    return response.data;
  },

  searchMessages: async (query: string, contactId?: string): Promise<Message[]> => {
    const params = new URLSearchParams({ q: query });
    if (contactId) params.append('contact_id', contactId);
    
    const response = await axiosInstance.get<Message[]>(`/messages/search?${params.toString()}`);
    return response.data;
  },

  uploadMedia: async (file: File): Promise<{ url: string; filename: string }> => {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await axiosInstance.post<{ url: string; filename: string }>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  updateMessageStatus: async (
    messageId: string,
    status: string
  ): Promise<void> => {
    await axiosInstance.patch(`/messages/${messageId}/status`, { status });
  },
};
