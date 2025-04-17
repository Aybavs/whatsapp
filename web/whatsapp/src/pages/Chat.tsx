import React, { useState } from 'react';

import { ChatHeader } from '@/components/chat/ChatHeader';
import { ContactList } from '@/components/chat/ContactList';
import { MessageInput } from '@/components/chat/MessageInput';
import { MessageList } from '@/components/chat/MessageList';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { useChat } from '@/hooks/useChat';
import { useContacts } from '@/hooks/useContacts';
import { useWebSocket } from '@/hooks/useWebSocket';
import { User } from '@/types';

const Chat: React.FC = () => {
  const { status: wsStatus } = useWebSocket();
  const {
    contacts,
    loading: contactsLoading,
    error: contactsError,
    searchUsers,
    addContact,
  } = useContacts();
  const [selectedContact, setSelectedContact] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<User[]>([]);

  const {
    messages,
    loading: messagesLoading,
    error: messagesError,
    sendMessage,
  } = useChat(selectedContact);

  const handleSelectContact = (contact: User) => {
    setSelectedContact(contact);
  };

  const handleSendMessage = (content: string) => {
    if (selectedContact) {
      sendMessage(content);
    }
  };

  const handleSearch = async () => {
    if (searchQuery.trim()) {
      const results = await searchUsers(searchQuery);
      setSearchResults(results);
    }
  };

  const handleAddContact = async (UserID: string) => {
    await addContact(UserID);
    setSearchResults([]);
    setSearchQuery("");
  };

  return (
    <AppLayout wsStatus={wsStatus}>
      <div className="flex flex-col md:flex-row w-full flex-1 overflow-hidden">
        <aside className="md:w-1/3 w-full max-h-[50vh] md:max-h-full flex flex-col border-b md:border-b-0 md:border-r border-gray-200 overflow-auto bg-white">
          <h2 className="p-4 font-semibold text-lg text-gray-800">
            Your Contacts
          </h2>
          <Separator />

          {contactsLoading ? (
            <div className="p-4 text-center text-gray-500">
              Loading contacts...
            </div>
          ) : contactsError ? (
            <div className="p-4 text-center text-red-500">
              Error: {contactsError}
            </div>
          ) : contacts && contacts.length > 0 ? (
            <ScrollArea className="flex-1">
              <ContactList
                contacts={contacts}
                selectedContact={selectedContact}
                onSelectContact={handleSelectContact}
                className="flex-1"
              />
            </ScrollArea>
          ) : (
            <div className="p-4 text-center text-gray-500">
              No contacts found. Search for users to add.
            </div>
          )}

          <div className="p-4 border-t">
            <h3 className="font-medium mb-2 text-gray-700">
              Find New Contacts
            </h3>
            <div className="flex gap-2 mb-2">
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by name or username"
              />
              <Button onClick={handleSearch}>Search</Button>
            </div>

            {searchResults.length > 0 && (
              <ul className="mt-2 border-t pt-2 space-y-1 max-h-40 overflow-y-auto text-sm">
                {searchResults.map((user) => (
                  <li
                    key={user.id}
                    className="flex justify-between items-center p-2 hover:bg-gray-100"
                  >
                    <span>{user.name || user.username}</span>
                    <Button
                      size="sm"
                      className="bg-whatsapp-green text-white"
                      onClick={() => handleAddContact(user.id)}
                    >
                      Add
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </aside>

        <section className="md:w-2/3 w-full flex flex-col flex-1 bg-gray-50">
          {selectedContact ? (
            <>
              <ChatHeader contact={selectedContact} />
              <div className="flex-1 overflow-y-auto">
                <MessageList
                  messages={messages}
                  loading={messagesLoading}
                  error={messagesError}
                />
              </div>
              <MessageInput
                onSendMessage={handleSendMessage}
                disabled={messagesLoading}
              />
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-gray-500 p-4">
              <div className="text-center">
                <p className="mb-2 text-lg font-medium">
                  Select a contact to start chatting
                </p>
                <p className="text-sm">
                  Your messages are end-to-end encrypted
                </p>
              </div>
            </div>
          )}
        </section>
      </div>
    </AppLayout>
  );
};

export default Chat;
