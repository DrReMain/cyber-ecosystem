import { Avatar, Card } from "antd";

interface ProfileIdentityProps {
  avatarUrl?: string;
  email: string;
  name: string;
}

export function ProfileIdentity({ avatarUrl, email, name }: Readonly<ProfileIdentityProps>) {
  return (
    <Card>
      <div className="flex flex-col items-center gap-3 py-2">
        <Avatar size={72} src={avatarUrl}>
          {name.charAt(0).toUpperCase()}
        </Avatar>
        <div className="flex min-w-0 flex-col items-center gap-0.5">
          <span className="truncate font-medium text-[15px]">{name}</span>
          <span className="truncate font-mono text-[12px] text-ink-tertiary">{email}</span>
        </div>
      </div>
    </Card>
  );
}
