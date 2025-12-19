import { PlusCircle, Users } from 'lucide-react';
import React, { memo, useCallback, useMemo, useState } from 'react';

import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Contact, Group, User } from '@/types';

interface ContactItemProps {
  contact: Contact;
  isSelected: boolean;
  onSelect: (contact: Contact) => void;
}

const ContactItem = memo(({ contact, isSelected, onSelect }: ContactItemProps) => {
  const handleClick = useCallback(() => {
    onSelect(contact);
  }, [contact, onSelect]);

  const isGroup = (c: Contact): c is Group => {
    return (c as Group).is_group === true;
  };

  const displayName = isGroup(contact) ? contact.name : contact.username;
  const avatarSrc = isGroup(contact) ? contact.avatar_url : (contact as User).avatar;
  const status = isGroup(contact) ? `${contact.member_ids.length} members` : (contact as User).status || "offline";

  return (
    <div
      onClick={handleClick}
      className={`flex items-center p-3 cursor-pointer transition-colors duration-150 rounded-md mx-2 mb-1
        ${
          isSelected
            ? "bg-green-50 border-l-4 border-green-400"
            : "hover:bg-gray-100"
        }
      `}
    >
      <div className="relative">
        <Avatar name={displayName} size="lg" src={avatarSrc} />
        {isGroup(contact) && (
          <div className="absolute -bottom-1 -right-1 bg-white rounded-full p-0.5">
           <Users className="w-3 h-3 text-gray-500" />
          </div>
        )}
      </div>
      <div className="ml-3 flex-1 overflow-hidden">
        <div className="flex justify-between items-baseline">
           <span className="font-medium text-sm">{displayName}</span>
           {/* Time placeholder if available */}
           {(contact as User).last_message_time && (
             <span className="text-xs text-gray-400">
               {new Date((contact as User).last_message_time!).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
             </span>
           )}
        </div>
        
        <div className="flex justify-between items-center">
            <div className="text-xs text-gray-500 truncate max-w-[140px]">
             {(contact as User).last_message ? (
                 <span className="text-gray-600">{(contact as User).last_message}</span>
             ) : (
                status
             )}
            </div>
        </div>
      </div>
    </div>
  );
});

ContactItem.displayName = 'ContactItem';

interface ContactListProps {
  contacts: User[];
  groups?: Group[];
  selectedContact: Contact | null;
  onSelectContact: (contact: Contact) => void;
  onAddContact?: () => void;
  onCreateGroup?: () => void;
  loading?: boolean;
  error?: string | null;
  className?: string;
}

const ContactListComponent: React.FC<ContactListProps> = ({
  contacts,
  groups = [],
  selectedContact,
  onSelectContact,
  onAddContact,
  onCreateGroup,
  loading = false,
  error = null,
  className = "",
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTab, setActiveTab] = useState<'all' | 'groups'>('all');

  // contacts array'ini memoize et
  const contactsArray = useMemo(() =>
    Array.isArray(contacts) ? contacts : [],
    [contacts]
  );
  
  const groupsArray = useMemo(() =>
    Array.isArray(groups) ? groups : [],
    [groups]
  );

  // Filtreleme işlemini memoize et
  const filteredItems = useMemo(() => {
    const allItems = activeTab === 'all' 
      ? [...groupsArray, ...contactsArray]
      : groupsArray;
      
    // Sort groups first, then users
    // But since we want chronological usually, let's keep it simple for now or specific sort
    // For now simple concat.
    
    return allItems.filter(
      (item: Contact) => {
        const name = (item as Group).name || (item as User).name || "";
        const username = (item as User).username || "";
        return name.toLowerCase().includes(searchQuery.toLowerCase()) ||
               username.toLowerCase().includes(searchQuery.toLowerCase());
      }
    );
  }, [contactsArray, groupsArray, searchQuery, activeTab]);

  const renderEmptyState = () => {
    if (searchQuery) {
      return (
        <div className="flex flex-col justify-center items-center h-40 text-gray-500 text-sm text-center px-4">
          <p className="mb-1">No results found</p>
          <p className="text-xs">Try a different search term</p>
        </div>
      );
    }

    return (
      <div className="flex flex-col justify-center items-center h-40 text-gray-500 text-sm px-4 text-center">
        <p className="mb-2">No conversations yet</p>
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
          placeholder="Search..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="flex items-center px-4 py-2 bg-white rounded-full shadow-sm border focus-within:ring-2 ring-green-500"
        />
      </div>
      
      {/* Tabs / Filter Buttons */}
      <div className="flex px-3 pb-2 gap-2">
         <button 
           onClick={() => setActiveTab('all')}
           className={`px-3 py-1 text-xs rounded-full transition-colors ${activeTab === 'all' ? 'bg-green-100 text-green-700 font-medium' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
         >
           All
         </button>
         <button 
           onClick={() => setActiveTab('groups')}
           className={`px-3 py-1 text-xs rounded-full transition-colors ${activeTab === 'groups' ? 'bg-green-100 text-green-700 font-medium' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
         >
           Groups
         </button>
         
         {onCreateGroup && (
           <button 
             onClick={onCreateGroup}
             className="ml-auto p-1 text-gray-500 hover:text-green-600 hover:bg-green-50 rounded"
             title="Create Group"
           >
             <Users className="w-4 h-4" />
           </button>
         )}
      </div>

      {/* Contact List */}
      <div className="flex-1 overflow-y-auto pr-1">
        {loading ? (
          <div className="flex justify-center items-center h-20 text-gray-500 text-sm">
            Loading...
          </div>
        ) : error ? (
          <div className="flex justify-center items-center h-20 text-red-500 text-sm">
            {error}
          </div>
        ) : filteredItems.length === 0 ? (
          renderEmptyState()
        ) : (
          filteredItems.map((item: Contact) => (
            <ContactItem
              key={item.id}
              contact={item}
              isSelected={selectedContact?.id === item.id}
              onSelect={onSelectContact}
            />
          ))
        )}
      </div>
    </div>
  );
};

// ContactList'i memo ile sarmalayarak export et
export const ContactList = memo(ContactListComponent);
