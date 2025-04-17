import { useState } from "react";

interface MessageInputProps {
  onSendMessage: (content: string, mediaUrl?: string) => void;
  disabled?: boolean;
}

export const MessageInput = ({
  onSendMessage,
  disabled,
}: MessageInputProps) => {
  const [message, setMessage] = useState("");
  const [isUploading] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() && !disabled && !isUploading) {
      onSendMessage(message);
      setMessage("");
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="p-4 border-t bg-white flex items-center gap-2"
    >
      <button
        type="button"
        className="p-2 rounded-full hover:bg-gray-100"
        title="Attach files"
        disabled={disabled || isUploading}
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

      <input
        type="text"
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        placeholder="Type a message"
        className="flex-1 p-2 border rounded-full focus:outline-none focus:ring-2 focus:ring-blue-300"
        disabled={disabled || isUploading}
      />

      <button
        type="submit"
        className="p-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 disabled:opacity-50 disabled:hover:bg-blue-500"
        disabled={!message.trim() || disabled || isUploading}
      >
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
      </button>
    </form>
  );
};
