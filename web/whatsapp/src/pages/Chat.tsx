import { Plus } from 'lucide-react';
import { useState } from 'react';

import { AddContactModal } from '@/components/chat/AddContactModal';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { ContactList } from '@/components/chat/ContactList';
import { MessageInput } from '@/components/chat/MessageInput';
import { MessageList } from '@/components/chat/MessageList';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/button';
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
  const [addContactOpen, setAddContactOpen] = useState(false);

  const {
    messages,
    loading: messagesLoading,
    error: messagesError,
    sendMessage,
  } = useChat(selectedContact);

  const handleSelectContact = (contact: User) => setSelectedContact(contact);

  const handleSendMessage = (content: string) => {
    if (selectedContact) sendMessage(content);
  };

  const handleAddContact = async (userId: string) => {
    const added = await addContact(userId);
    if (!added) {
      // kullanıcıya toast göster vs.
      console.log("Already added.");
    }
  };

  return (
    <AppLayout wsStatus={wsStatus}>
      <div className="flex flex-col md:flex-row w-full flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside className="md:w-1/3 w-full max-h-[50vh] md:max-h-full flex flex-col border-b md:border-b-0 md:border-r border-gray-200 bg-white">
          <div className="p-4 flex items-center justify-between">
            <h2 className="font-semibold text-lg text-gray-800">
              Your Contacts
            </h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setAddContactOpen(true)}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
          <Separator />

          {/* Contact List */}
          {contactsLoading ? (
            <div className="p-4 text-center text-gray-500">
              Loading contacts...
            </div>
          ) : contactsError ? (
            <div className="p-4 text-center text-red-500">
              Error: {contactsError}
            </div>
          ) : contacts?.length > 0 ? (
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
              No contacts found. Click + to add new.
            </div>
          )}
        </aside>

        {/* Chat Area */}
        <section className="md:w-2/3 w-full flex flex-col flex-1 h-[calc(100vh-64px)] bg-gray-50">
          {selectedContact ? (
            <>
              <ChatHeader contact={selectedContact} />
              {/* Change overflow-hidden to overflow-auto */}
              <div className="flex-1 overflow-auto">
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

      {/* Modal */}
      <AddContactModal
        open={addContactOpen}
        onOpenChange={setAddContactOpen}
        onSearch={searchUsers}
        onAddContact={handleAddContact}
        existingContacts={contacts}
      />
    </AppLayout>
  );
};

export default Chat;
