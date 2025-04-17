import { useCallback, useEffect, useState } from 'react';

import { userApi } from '@/api/userApi';
import { useAuth } from '@/hooks/useAuth';
import { User } from '@/types';

export const useContacts = () => {
  const [contacts, setContacts] = useState<User[]>([]);
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { user } = useAuth();

  // Fetch actual contacts (people the user has chatted with)
  const fetchContacts = useCallback(async () => {
    if (!user) return;

    setLoading(true);
    setError(null);
    try {
      // This should call an endpoint that returns actual contacts
      const contacts = await userApi.getContacts();
      setContacts(contacts);
    } catch (err) {
      console.error("Failed to fetch contacts:", err);
      setError("Failed to load contacts");
    } finally {
      setLoading(false);
    }
  }, [user]);

  // Search for users (for starting new conversations)
  const searchUsers = useCallback(
    async (searchQuery: string) => {
      if (!user || !searchQuery.trim()) return [];

      setLoading(true);
      try {
        const response = await userApi.searchUsers(searchQuery);

        // Filter out current user from search results
        const filteredUsers = response.filter(
          (searchedUser: User) => searchedUser.id !== user.id
        );

        setAllUsers(filteredUsers);
        return filteredUsers;
      } catch (err) {
        console.error("Failed to search users:", err);
        setError("Failed to search users");
        return [];
      } finally {
        setLoading(false);
      }
    },
    [user]
  );

  // Add a user as a contact
  const addContact = useCallback(
    async (contactId: string) => {
      if (!user) return false;

      try {
        await userApi.addContact(contactId);
        // Refresh contacts list after adding
        fetchContacts();
        return true;
      } catch (err) {
        console.error("Failed to add contact:", err);
        setError("Failed to add contact");
        return false;
      }
    },
    [user, fetchContacts]
  );

  // Fetch contacts on mount
  useEffect(() => {
    if (user) {
      fetchContacts();
    }
  }, [user, fetchContacts]);

  return {
    // Existing contacts the user has
    contacts,
    // For searching all users
    allUsers,
    loading,
    error,
    // Methods
    fetchContacts,
    searchUsers,
    addContact,
  };
};
