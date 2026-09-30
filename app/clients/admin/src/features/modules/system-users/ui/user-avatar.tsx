import { Avatar } from "antd";

const AVATAR_COLORS = ["#1677ff", "#722ed1", "#13c2c2", "#fa8c16", "#eb2f96", "#52c41a"];

function avatarColor(seed: string): string {
  let sum = 0;
  for (const ch of seed) sum += ch.charCodeAt(0);
  return AVATAR_COLORS[sum % AVATAR_COLORS.length] ?? "#1677ff";
}

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
