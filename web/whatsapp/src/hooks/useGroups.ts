import { useState, useEffect, useCallback } from 'react';
import { groupApi } from '@/api/groupApi';
import { Group } from '@/types';

export const useGroups = () => {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchGroups = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await groupApi.getUserGroups();
      setGroups(data);
    } catch (err) {
      setError("Failed to load groups");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGroups();
  }, [fetchGroups]);

  const refreshGroups = useCallback(() => {
    fetchGroups();
  }, [fetchGroups]);

  return {
    groups,
    loading,
    error,
    refreshGroups
  };
};
