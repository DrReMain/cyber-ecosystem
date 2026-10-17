import { Avatar } from "antd";

import { avatarColor } from "#/libs";

export function localPart(email: string): string {
  return email.split("@")[0] ?? email;
}

interface UserAvatarProps {
  url: string | undefined;
  email: string;
  size?: number;
}

export function UserAvatar({ url, email, size }: Readonly<UserAvatarProps>) {
  return (
    <Avatar size={size} src={url} style={{ backgroundColor: avatarColor(email), flexShrink: 0 }}>
      {localPart(email).charAt(0).toUpperCase()}
    </Avatar>
  );
}
