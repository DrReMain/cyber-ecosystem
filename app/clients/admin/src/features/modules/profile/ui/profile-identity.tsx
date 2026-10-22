import { Avatar, Card } from "antd";

interface ProfileIdentityProps {
  avatarUrl?: string;
  account: string;
  name: string;
}

export function ProfileIdentity({ avatarUrl, account, name }: Readonly<ProfileIdentityProps>) {
  return (
    <Card>
      <div className="flex flex-col items-center gap-3 py-2">
        <Avatar size={72} src={avatarUrl}>
          {name.charAt(0).toUpperCase()}
        </Avatar>
        <div className="flex min-w-0 flex-col items-center gap-0.5">
          <span className="truncate text-[15px] font-medium">{name}</span>
          <span className="text-ink-tertiary truncate font-mono text-[12px]">{account}</span>
        </div>
      </div>
    </Card>
  );
}
