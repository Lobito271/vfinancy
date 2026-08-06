import { wailsClient } from '../bindings';
import type { Role } from '@/constants/permissions';

export interface LoginInput {
  username: string;
  password: string;
  remember?: boolean;
}

export interface LoginResult {
  token: string;
  expiresAt: string;
  user: {
    id: string;
    fullName: string;
    email: string;
    username: string;
    roles: Role[];
    company: string;
    mustChangePassword: boolean;
  };
}

export const authService = {
  async login(input: LoginInput): Promise<LoginResult> {
    const res = await wailsClient.login(input.username, input.password, input.remember ?? false);
    return {
      token: res.token,
      expiresAt: res.expiresAt,
      user: {
        id: res.userId,
        fullName: res.fullName,
        email: res.email,
        username: res.username,
        roles: res.roles as Role[],
        company: res.companyId,
        mustChangePassword: res.mustChangePassword,
      },
    };
  },
  async logout(token: string): Promise<void> {
    await wailsClient.logout(token);
  },
  async changePassword(currentPassword: string, newPassword: string, token: string): Promise<void> {
    await wailsClient.changePassword(currentPassword, newPassword, token);
  },
  async validateSession(token: string): Promise<boolean> {
    return wailsClient.validateSession(token);
  },
};
