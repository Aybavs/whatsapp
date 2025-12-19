import { X } from 'lucide-react';
import React, { useState } from 'react';

import { groupApi } from '@/api/groupApi';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { User } from '@/types';

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  contacts: User[];
  onGroupCreated: () => void;
}

export const CreateGroupModal: React.FC<CreateGroupModalProps> = ({
  isOpen,
  onClose,
  contacts,
  onGroupCreated,
}) => {
  const [groupName, setGroupName] = useState("");
  const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>([]);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!groupName.trim()) {
        setError("Group name is required");
        return;
    }
    if (selectedMemberIds.length === 0) {
        setError("Select at least one member");
        return;
    }

    setCreating(true);
    setError(null);
    try {
      await groupApi.createGroup({
        name: groupName,
        member_ids: selectedMemberIds,
      });
      onGroupCreated();
      onClose();
    } catch (err) {
      console.error("Failed to create group:", err);
      setError("Failed to create group");
    } finally {
      setCreating(false);
    }
  };

  const toggleMember = (userId: string) => {
    setSelectedMemberIds(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-lg w-full max-w-md shadow-xl flex flex-col max-h-[90vh]">
        <div className="flex justify-between items-center p-4 border-b">
          <h2 className="text-xl font-semibold">Create New Group</h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded-full">
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        <form onSubmit={handleCreate} className="flex-1 overflow-hidden flex flex-col">
          <div className="p-4 space-y-4 overflow-y-auto">
            {error && (
              <div className="p-3 bg-red-100 text-red-600 rounded-md text-sm">
                {error}
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Group Name
              </label>
              <Input
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
                placeholder="Enter group name"
                className="w-full"
                autoFocus
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select Members
              </label>
              <div className="space-y-2 border rounded-md p-2 max-h-60 overflow-y-auto">
                {contacts.length === 0 ? (
                    <p className="text-sm text-gray-500 text-center p-2">No contacts available</p>
                ) : (
                    contacts.map(contact => (
                    <div 
                        key={contact.id} 
                        className={`flex items-center p-2 rounded cursor-pointer ${selectedMemberIds.includes(contact.id) ? 'bg-green-50' : 'hover:bg-gray-50'}`}
                        onClick={() => toggleMember(contact.id)}
                    >
                        <div className={`w-4 h-4 rounded border mr-3 flex items-center justify-center ${selectedMemberIds.includes(contact.id) ? 'bg-green-500 border-green-500' : 'border-gray-300'}`}>
                        {selectedMemberIds.includes(contact.id) && (
                            <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
                            </svg>
                        )}
                        </div>
                        <span className="text-sm font-medium">{contact.username}</span>
                    </div>
                    ))
                )}
              </div>
              <p className="text-xs text-gray-500 mt-1">
                {selectedMemberIds.length} members selected
              </p>
            </div>
          </div>

          <div className="p-4 border-t flex justify-end gap-2 bg-gray-50">
            <Button type="button" variant="outline" onClick={onClose} disabled={creating}>
              Cancel
            </Button>
            <Button type="submit" disabled={creating || !groupName.trim() || selectedMemberIds.length === 0} className="bg-green-600 hover:bg-green-700 text-white">
              {creating ? 'Creating...' : 'Create Group'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
