import { Loader2, Plus, Search } from 'lucide-react';
import { useState } from 'react';

import { Avatar } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { User } from '@/types';

interface AddContactModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSearch: (query: string) => Promise<User[]>;
  onAddContact: (userId: string) => Promise<void>;
  existingContacts: User[];
}

export const AddContactModal: React.FC<AddContactModalProps> = ({
  open,
  onOpenChange,
  onSearch,
  onAddContact,
  existingContacts,
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [results, setResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;

    setIsSearching(true);
    try {
      const res = await onSearch(searchQuery);
      const existingIds = new Set(existingContacts.map((c) => c.id));
      const filtered = res.filter((u) => !existingIds.has(u.id));

      setResults(filtered);
      setHasSearched(true);
    } catch (error) {
      console.error("Search failed", error);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold text-gray-800">
            Find New Contacts
          </DialogTitle>
        </DialogHeader>

        {/* Improved Search bar */}
        <div className="mt-4">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleSearch();
            }}
            className="relative"
          >
            <div className="relative rounded-full overflow-hidden border shadow-sm focus-within:ring-2 ring-green-500">
              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                <Search size={18} />
              </div>
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by name or username"
                className="w-full h-11 pl-10 pr-24 border-none focus-visible:ring-0 text-sm"
              />
              <Button
                type="submit"
                disabled={isSearching || !searchQuery.trim()}
                className="absolute right-1 top-1/2 -translate-y-1/2 h-9 px-4 text-sm font-medium rounded-full bg-green-600 text-white hover:bg-green-700"
              >
                {isSearching ? (
                  <Loader2 size={16} className="animate-spin mr-1" />
                ) : (
                  "Search"
                )}
              </Button>
            </div>
          </form>
        </div>

        {/* Results with improved UI */}
        {hasSearched && (
          <div className="mt-6">
            <h3 className="text-sm font-medium text-gray-500 mb-3">
              {results.length === 0
                ? "No users found. Try a different search term."
                : `Found ${results.length} user${
                    results.length !== 1 ? "s" : ""
                  }`}
            </h3>

            {results.length > 0 && (
              <ScrollArea className="max-h-72 pr-2">
                <div className="space-y-3">
                  {results.map((user) => (
                    <div
                      key={user.id}
                      className="flex items-center justify-between p-3 rounded-lg bg-white border border-gray-100 hover:shadow-md transition-shadow"
                    >
                      <div className="flex items-center gap-3">
                        <Avatar
                          name={user.name || user.username}
                          className="border-2 border-gray-100"
                        />
                        <div>
                          <div className="text-sm font-medium text-gray-800">
                            {user.name || user.username}
                          </div>
                          {user.name && user.username && (
                            <div className="text-xs text-gray-500">
                              @{user.username}
                            </div>
                          )}
                        </div>
                      </div>
                      <Button
                        size="sm"
                        className="rounded-full bg-green-600 text-white hover:bg-green-700"
                        onClick={() => onAddContact(user.id)}
                      >
                        <Plus size={16} className="mr-1" />
                        Add
                      </Button>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
};
