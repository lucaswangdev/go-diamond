// Simple auth service - in production, this should be a server-side endpoint
// For demo purposes, we use a static email/password check and generate a mock token

interface AuthUser {
  email: string;
  name: string;
}

const MOCK_USERS: Record<string, string> = {
  'admin@example.com': 'admin123',
  'user@example.com': 'user123',
};

class AuthService {
  async login(email: string, password: string): Promise<{ token: string; user: AuthUser }> {
    // Simulate API delay
    await new Promise((resolve) => setTimeout(resolve, 500));

    const validPassword = MOCK_USERS[email];
    if (!validPassword || validPassword !== password) {
      throw new Error('Invalid email or password');
    }

    const token = btoa(`${email}:${Date.now()}`);
    const user: AuthUser = { email, name: email.split('@')[0] };

    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));

    return { token, user };
  }

  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  getUser(): AuthUser | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }
}

export const authService = new AuthService();
export default authService;