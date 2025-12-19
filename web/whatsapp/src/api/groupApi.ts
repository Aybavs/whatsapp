import axiosInstance from './axios';
import { Group } from '../types';

export interface CreateGroupRequest {
  name: string;
  description?: string;
  member_ids: string[];
}

export const groupApi = {
  createGroup: async (data: CreateGroupRequest): Promise<Group> => {
    const response = await axiosInstance.post<Group>('/groups', data);
    return { ...response.data, is_group: true };
  },

  getUserGroups: async (): Promise<Group[]> => {
    const response = await axiosInstance.get<Group[]>('/groups');
    return response.data.map(g => ({ ...g, is_group: true }));
  },
};
