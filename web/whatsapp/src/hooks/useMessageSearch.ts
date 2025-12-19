import { useState, useCallback, useEffect } from 'react';
import { messageApi } from '../api/messageApi';
import { Message } from '../types';
import { useAuth } from './useAuth';

export const useMessageSearch = (contactId?: string) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Message[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { isAuthenticated } = useAuth();
  
  const search = useCallback(async (searchQuery: string) => {
    if (!searchQuery.trim() || !isAuthenticated) {
      setResults([]);
      return;
    }

    setIsSearching(true);
    setError(null);

    try {
      const messages = await messageApi.searchMessages(searchQuery, contactId);
      setResults(messages);
    } catch (err) {
      console.error('Failed to search messages:', err);
      setError('Failed to search messages');
      setResults([]);
    } finally {
      setIsSearching(false);
    }
  }, [contactId, isAuthenticated]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (query) {
        search(query);
      } else {
        setResults([]);
      }
    }, 500); // Debounce search

    return () => clearTimeout(timeoutId);
  }, [query, search]);

  return {
    query,
    setQuery,
    results,
    isSearching,
    error,
    search
  };
};
