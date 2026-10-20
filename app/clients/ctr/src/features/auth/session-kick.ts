import { createChannel } from "#/libs";

const kickChannel = () => createChannel<number>("session-kick:ctr");

export function broadcastSessionKick(): void {
  const channel = kickChannel();
  channel.publish(Date.now());
  channel.close();
}

export function subscribeSessionKick(handler: () => void): () => void {
  const channel = kickChannel();
  const unsubscribe = channel.subscribe(() => handler());
  return () => {
    unsubscribe();
    channel.close();
  };
}
