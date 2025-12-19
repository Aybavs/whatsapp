import { Plus } from 'lucide-react';
import { useCallback, useState, useEffect } from 'react';

import { AddContactModal } from '@/components/chat/AddContactModal';
import { CreateGroupModal } from '@/components/chat/CreateGroupModal';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { ContactList } from '@/components/chat/ContactList';
import { MessageInput } from '@/components/chat/MessageInput';
import { MessageList } from '@/components/chat/MessageList';
import { SearchPanel } from '@/components/chat/SearchPanel';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { useAuth } from '@/hooks/useAuth';
import { useChat } from '@/hooks/useChat';
import { useContacts } from '@/hooks/useContacts';
import { useGroups } from '@/hooks/useGroups';
import { useTypingIndicator } from '@/hooks/useTypingIndicator';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Contact } from '@/types';

const Chat: React.FC = () => {
  const { status: wsStatus, onMessage } = useWebSocket();
  const {
    contacts,
    loading: contactsLoading,
    error: contactsError,
    searchUsers,
    addContact,
    updateContactLastMessage, // New function from useContacts
  } = useContacts();

  const {
    groups,
    loading: groupsLoading,
    refreshGroups
  } = useGroups();

  const { user } = useAuth();
  const [selectedContact, setSelectedContact] = useState<Contact | null>(null);
  const [addContactOpen, setAddContactOpen] = useState(false);
  const [createGroupOpen, setCreateGroupOpen] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  const {
    messages,
    loading: messagesLoading,
    error: messagesError,
    sendMessage,
  } = useChat(selectedContact);

  // Typing indicator
  const { isTyping, sendTyping } = useTypingIndicator(selectedContact?.id || null);

  // Memoized event handlers
  const handleSelectContact = useCallback((contact: Contact) => {
    setSelectedContact(contact);
    setIsSearchOpen(false); // Close search when changing contact
  }, []);

  const handleSendMessage = useCallback(async (content: string, mediaUrl?: string) => {
    await sendMessage(content, mediaUrl);
    // Update sidebar for the sent message
    if (selectedContact) {
      updateContactLastMessage(
        selectedContact.id, 
        content, 
        new Date().toISOString()
      );
    }
  }, [sendMessage, selectedContact, updateContactLastMessage]);

  // Listen for incoming messages to update sidebar
// Listen for incoming messages to update sidebar
  useEffect(() => {
    const unsubscribe = onMessage((message: any) => {
        // If it's a regular message
        if (message.content) {
            // Determine who the contact is. 
            // If we received it, the contact is the sender.
            // If we sent it (echo), the contact is the receiver.
            // Assuming this is mostly for incoming messages:
            const contactId = message.sender_id;
            // Also need to handle group messages if we support them later
            
            updateContactLastMessage(
                contactId,
                message.content,
                message.created_at || new Date().toISOString()
            );
        }
    });
    return () => {
        unsubscribe();
    };
  }, [onMessage, updateContactLastMessage]);

  const handleAddContact = useCallback(async (userId: string) => {
    const added = await addContact(userId);
    if (!added) {
      console.log("Already added.");
    }
  }, [addContact]);

  const handleOpenAddContact = useCallback(() => {
    setAddContactOpen(true);
  }, []);

  const handleSearchToggle = useCallback(() => {
    setIsSearchOpen((prev) => !prev);
  }, []);

  const handleGroupCreated = () => {
      refreshGroups();
  };

  return (
    <AppLayout wsStatus={wsStatus}>
      <div className="flex flex-col md:flex-row w-full flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside className="md:w-1/3 w-full max-h-[50vh] md:max-h-full flex flex-col border-b md:border-b-0 md:border-r border-gray-200 bg-white">
          <div className="p-4 flex items-center justify-between">
            <h2 className="font-semibold text-lg text-gray-800">
              Chats
            </h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleOpenAddContact}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>
          <Separator />

          {/* Contact List */}
          {contactsLoading || groupsLoading ? (
            <div className="p-4 text-center text-gray-500">
              Loading...
            </div>
          ) : contactsError ? (
            <div className="p-4 text-center text-red-500">
              Error: {contactsError}
            </div>
          ) : (
            <ScrollArea className="flex-1">
              <ContactList
                contacts={contacts}
                groups={groups}
                selectedContact={selectedContact}
                onSelectContact={handleSelectContact}
                onAddContact={handleOpenAddContact}
                onCreateGroup={() => setCreateGroupOpen(true)}
                className="flex-1"
              />
            </ScrollArea>
          )}
        </aside>

        {/* Chat Area */}
        <section className="md:w-2/3 w-full flex flex-col flex-1 h-[calc(100vh-64px)] bg-gray-50 relative">
          {selectedContact ? (
            <>
              <ChatHeader 
                contact={selectedContact} 
                isTyping={isTyping} 
                onSearchClick={handleSearchToggle}
              />
              {/* Change overflow-hidden to overflow-auto */}
              <div className="flex-1 overflow-auto relative">
                <MessageList
                  messages={messages}
                  loading={messagesLoading}
                  error={messagesError}
                />
                
                {isSearchOpen && user && (
                  <SearchPanel
                    onClose={() => setIsSearchOpen(false)}
                    currentUserId={user.id}
                    contactId={selectedContact.id}
                  />
                )}
              </div>
              <MessageInput
                onSendMessage={handleSendMessage}
                onTyping={sendTyping}
                disabled={messagesLoading}
              />
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-gray-500 p-4">
              <div className="text-center">
                <p className="mb-2 text-lg font-medium">
                  Select a contact or group to start chatting
                </p>
                <p className="text-sm">
                  Your messages are end-to-end encrypted
                </p>
              </div>
            </div>
          )}
        </section>
      </div>

      {/* Modals */}
      <AddContactModal
        open={addContactOpen}
        onOpenChange={setAddContactOpen}
        onSearch={searchUsers}
        onAddContact={handleAddContact}
        existingContacts={contacts}
      />
      
      <CreateGroupModal
        isOpen={createGroupOpen}
        onClose={() => setCreateGroupOpen(false)}
        contacts={contacts}
        onGroupCreated={handleGroupCreated}
      />
    </AppLayout>
  );
};

export default Chat;
