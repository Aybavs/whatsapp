import axiosInstance from '@/api/axios';
import { Message } from '@/types';

export const messageApi = {
  getMessages: async (UserID: string): Promise<Message[]> => {
    const response = await axiosInstance.get<Message[]>(`/messages/${UserID}`);
    return response.data;
  },

  sendMessage: async (
    receiverId: string,
    content: string,
    mediaUrl?: string
  ): Promise<Message> => {
    const response = await axiosInstance.post<Message>("/messages", {
      receiver_id: receiverId,
      content,
      media_url: mediaUrl,
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
