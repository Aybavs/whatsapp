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
  onAddContact?: () => void;
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

  const contactsArray = Array.isArray(contacts) ? contacts : [];

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
        <div className="flex flex-col justify-center items-center h-40 text-gray-500 text-sm text-center px-4">
          <p className="mb-1">No contacts found</p>
          <p className="text-xs">Try a different search term</p>
        </div>
      );
    }

    return (
      <div className="flex flex-col justify-center items-center h-40 text-gray-500 text-sm px-4 text-center">
        <p className="mb-2">You have no contacts yet</p>
        {onAddContact && (
          <Button
            onClick={onAddContact}
            variant="outline"
            className="mt-2 text-sm text-green-600 hover:bg-green-50"
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
      {/* Search */}
      <div className="p-3">
        <Input
          placeholder="Search contacts..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex items-center px-4 py-2 bg-white rounded-full shadow-sm border focus-within:ring-2 ring-green-500"
        />
      </div>

      {/* Contact List */}
      <div className="flex-1 overflow-y-auto pr-1">
        {loading ? (
          <div className="flex justify-center items-center h-20 text-gray-500 text-sm">
            Loading contacts...
          </div>
        ) : error ? (
          <div className="flex justify-center items-center h-20 text-red-500 text-sm">
            {error}
          </div>
        ) : !contacts || filteredContacts.length === 0 ? (
          renderEmptyState()
        ) : (
          filteredContacts.map((contact) => (
            <div
              key={contact.id}
              onClick={() => onSelectContact(contact)}
              className={`flex items-center p-3 cursor-pointer transition-colors duration-150 rounded-md mx-2 mb-1
                ${
                  selectedContact?.id === contact.id
                    ? "bg-green-50 border-l-4 border-green-400"
                    : "hover:bg-gray-100"
                }
              `}
            >
              <Avatar name={contact.username} size="lg" src={contact.avatar} />
              <div className="ml-3">
                <div className="font-medium text-sm">{contact.username}</div>
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
