import * as React from "react";
import * as AvatarPrimitive from "@radix-ui/react-avatar";

import { cn } from "@/lib/utils";

type AvatarSize = "xs" | "sm" | "md" | "lg" | "xl" | "2xl";

interface AvatarProps
  extends React.ComponentProps<typeof AvatarPrimitive.Root> {
  name?: string;
  src?: string;
  size?: AvatarSize;
}

function Avatar({
  className,
  name,
  src,
  size = "md",
  children,
  ...props
}: AvatarProps) {
  // If both name and children are provided, children take precedence
  // If no children are provided but name exists, we'll generate a fallback with initials
  const hasCustomContent = React.Children.count(children) > 0;

  // Map sizes to tailwind classes
  const sizeClasses = {
    xs: "size-6",
    sm: "size-8",
    md: "size-10",
    lg: "size-12",
    xl: "size-16",
    "2xl": "size-20",
  };

  return (
    <AvatarPrimitive.Root
      data-slot="avatar"
      className={cn(
        "relative flex shrink-0 overflow-hidden rounded-full",
        sizeClasses[size],
        className
      )}
      {...props}
    >
      {hasCustomContent ? (
        children
      ) : (
        <>
          {src && <AvatarImage src={src} alt={name || "Avatar"} />}
          {name && <AvatarFallback>{getInitials(name)}</AvatarFallback>}
        </>
      )}
    </AvatarPrimitive.Root>
  );
}

function AvatarImage({
  className,
  ...props
}: React.ComponentProps<typeof AvatarPrimitive.Image>) {
  return (
    <AvatarPrimitive.Image
      data-slot="avatar-image"
      className={cn("aspect-square size-full", className)}
      {...props}
    />
  );
}

function AvatarFallback({
  className,
  ...props
}: React.ComponentProps<typeof AvatarPrimitive.Fallback>) {
  return (
    <AvatarPrimitive.Fallback
      data-slot="avatar-fallback"
      className={cn(
        "bg-muted flex size-full items-center justify-center rounded-full",
        className
      )}
      {...props}
    />
  );
}

// Helper function to get initials from a name
function getInitials(name: string): string {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

export { Avatar, AvatarImage, AvatarFallback };
