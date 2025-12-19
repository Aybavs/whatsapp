import { messageApi } from '@/api/messageApi';
import { useCallback, useRef, useState } from 'react';

interface MessageInputProps {
  onSendMessage: (content: string, mediaUrl?: string) => void;
  onTyping?: (isTyping: boolean) => void;
  disabled?: boolean;
}

export const MessageInput = ({
  onSendMessage,
  onTyping,
  disabled,
}: MessageInputProps) => {
  const [message, setMessage] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setMessage(value);

    // Send typing event
    if (onTyping) {
      onTyping(value.length > 0);
    }
  }, [onTyping]);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    try {
      const { url } = await messageApi.uploadMedia(file);
      // Send message with media immediately
      onSendMessage(message || "Sent an attachment", url); 
      setMessage("");
      // Stop typing indicator
      if (onTyping) onTyping(false);
    } catch (error) {
      console.error("Failed to upload media:", error);
      alert("Failed to upload media");
    } finally {
      setIsUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() && !disabled && !isUploading) {
      onSendMessage(message);
      setMessage("");

      // Stop typing indicator when message is sent
      if (onTyping) {
        onTyping(false);
      }
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="p-3 border-t bg-white flex items-center gap-2"
    >
      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        onChange={handleFileSelect}
        accept="image/*,video/*,application/pdf"
      />
      
      {/* Attach Button */}
      <button
        type="button"
        className="p-2 rounded-full hover:bg-gray-100 disabled:opacity-40"
        title="Attach file"
        disabled={disabled || isUploading}
        onClick={() => fileInputRef.current?.click()}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-5 w-5 text-gray-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
          />
        </svg>
      </button>

      {/* Input */}
      <input
        type="text"
        value={message}
        onChange={handleChange}
        placeholder={isUploading ? "Uploading..." : "Type a message"}
        className="flex-1 bg-gray-100 rounded-full px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-300"
        disabled={disabled || isUploading}
      />

      {/* Send Button */}
      <button
        type="submit"
        className="p-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 disabled:opacity-50"
        disabled={!message.trim() || disabled || isUploading}
        title="Send"
      >
        {isUploading ? (
         <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        ) : (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
            />
          </svg>
        )}
      </button>
    </form>
  );
};
