import { Avatar } from "antd";

import { avatarColor } from "#/libs";

export function displayName(account: string): string {
  return account.split("@")[0] ?? account;
}

interface UserAvatarProps {
  url: string | undefined;
  account: string;
  size?: number;
}

export function UserAvatar({ url, account, size }: Readonly<UserAvatarProps>) {
  return (
    <Avatar size={size} src={url} style={{ backgroundColor: avatarColor(account), flexShrink: 0 }}>
      {displayName(account).charAt(0).toUpperCase()}
    </Avatar>
  );
}
