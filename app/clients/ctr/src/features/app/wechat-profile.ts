// The OAuth moment is the only moment wechat's nickname/avatar exist for
// us — the userinfo access token is spent on arrival. The callback pages
// park what they learned here so the nickname step can prefill from it;
// the scope survives wizard navigation and page reloads within the tab.
// It does NOT police itself: storage lives until the tab closes, so the
// lifecycle is owned by the callers — sign-out, a refused bind, and the
// nickname step's successful write each clear it. Without those, a later
// account in the same tab inherits the previous seat's wechat identity.

const STORAGE_KEY = "ctr:wechat-profile";

export interface WechatProfile {
  nickname: string;
  avatarUrl: string;
}

export function parkWechatProfile(profile: WechatProfile): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(profile));
  } catch {
    // Private-mode quota refusals degrade to no prefill, never to a crash.
  }
}

export function clearWechatProfile(): void {
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // Private-mode storage refusals degrade the same way as parking one.
  }
}

export function readWechatProfile(): WechatProfile | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (raw === null) {
      return null;
    }
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed !== "object" ||
      parsed === null ||
      typeof (parsed as WechatProfile).nickname !== "string" ||
      typeof (parsed as WechatProfile).avatarUrl !== "string"
    ) {
      return null;
    }
    return parsed as WechatProfile;
  } catch {
    return null;
  }
}
