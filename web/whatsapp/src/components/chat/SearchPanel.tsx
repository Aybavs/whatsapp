import React from 'react';
import { useMessageSearch } from '../../hooks/useMessageSearch';
import { MessageBubble } from './MessageBubble';
import { Message } from '@/types';
import { X } from 'lucide-react';

interface SearchPanelProps {
  onClose: () => void;
  currentUserId: string;
  contactId?: string;
}

export const SearchPanel: React.FC<SearchPanelProps> = ({ onClose, currentUserId, contactId }) => {
  const { query, setQuery, results, isSearching } = useMessageSearch(contactId);

  return (
    <div className="w-96 border-l bg-gray-50 flex flex-col h-full absolute right-0 top-0 z-20 shadow-lg animate-slide-in">
      <div className="h-16 px-4 bg-white border-b flex items-center gap-3">
        <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full">
          <X className="w-5 h-5 text-gray-500" />
        </button>
        <div className="text-gray-700 font-medium">Search Messages</div>
      </div>

      <div className="p-4 bg-white border-b">
        <div className="relative">
          <input
            type="text"
            placeholder="Search..."
            className="w-full pl-4 pr-10 py-2 bg-gray-100 rounded-lg focus:outline-none focus:ring-1 focus:ring-green-500"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {isSearching && (
          <div className="text-center text-gray-500 text-sm mt-4">Searching...</div>
        )}
        
        {!isSearching && query && results.length === 0 && (
          <div className="text-center text-gray-500 text-sm mt-4">No messages found</div>
        )}

        {results.map((message: Message) => (
          <div key={message.id} className="cursor-pointer hover:bg-gray-100 p-2 rounded-lg transition-colors">
            <div className="text-xs text-gray-400 mb-1">
              {new Date(message.created_at).toLocaleDateString()}
            </div>
            <MessageBubble
              message={message}
              isSender={message.sender_id === currentUserId}
            />
          </div>
        ))}
      </div>
    </div>
  );
};
