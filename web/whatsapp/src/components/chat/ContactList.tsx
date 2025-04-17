import { PlusCircle } from 'lucide-react';
import React, { useState } from 'react';

import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { User } from '@/types';

interface ContactListProps {
  contacts: User[];
  selectedContact: User | null;
  onSelectContact: (contact: User) => void;
  onAddContact?: () => void; // Yeni contact ekleme fonksiyonu
  loading?: boolean;
  error?: string | null;
  className?: string;
}

export const ContactList: React.FC<ContactListProps> = ({
  contacts,
  selectedContact,
  onSelectContact,
  onAddContact,
  loading = false,
  error = null,
  className = "",
}) => {
  const [searchQuery, setSearchQuery] = useState("");

  // Ensure contacts is an array before filtering
  const contactsArray = Array.isArray(contacts) ? contacts : [];

  // Filter contacts by search query
  const filteredContacts = contactsArray.filter(
    (contact) =>
      (contact.name?.toLowerCase() || "").includes(searchQuery.toLowerCase()) ||
      (contact.username?.toLowerCase() || "").includes(
        searchQuery.toLowerCase()
      )
  );

  const renderEmptyState = () => {
    if (searchQuery) {
      return (
        <div className="flex flex-col justify-center items-center h-40 text-gray-500">
          <p className="mb-2">No contacts found</p>
          <p className="text-sm">Try a different search term</p>
        </div>
      );
    }

    return (
      <div className="flex flex-col justify-center items-center h-40 text-gray-500">
        <p className="mb-2">You have no contacts yet</p>
        {onAddContact && (
          <Button
            onClick={onAddContact}
            variant="ghost"
            className="flex items-center mt-2 text-primary"
          >
            <PlusCircle className="w-4 h-4 mr-2" />
            Add New Contact
          </Button>
        )}
      </div>
    );
  };

  return (
    <div className={`flex flex-col h-full ${className}`}>
      <div className="p-3">
        <Input
          placeholder="Search contacts..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="py-1.5 bg-gray-100"
        />
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="flex justify-center items-center h-20 text-gray-500">
            Loading contacts...
          </div>
        ) : error ? (
          <div className="flex justify-center items-center h-20 text-red-500">
            {error}
          </div>
        ) : !contacts || filteredContacts.length === 0 ? (
          renderEmptyState()
        ) : (
          filteredContacts.map((contact) => (
            <div
              key={contact.id}
              className={`flex items-center p-3 cursor-pointer hover:bg-gray-100 ${
                selectedContact?.id === contact.id ? "bg-gray-200" : ""
              }`}
              onClick={() => onSelectContact(contact)}
            >
              <Avatar name={contact.username} size="lg" src={contact.avatar} />
              <div className="ml-3">
                <div className="font-medium">{contact.username}</div>
                <div className="text-xs text-gray-500">
                  {contact.status || "offline"}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
