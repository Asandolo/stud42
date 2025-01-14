import { signIn, useSession } from 'next-auth/react';
import { useEffect } from 'react';
import type { ComponentWithAuthGuard } from './types';

type AuthGuardAuth = {
  loading: JSX.Element;
  required: boolean;
  unauthorized: Function;
};

export function AuthGuard({
  children,
}: {
  children: JSX.Element & Partial<ComponentWithAuthGuard>;
}) {
  const { data: session, status } = useSession();
  const {
    loading = <></>,
    required = true,
    unauthorized = signIn,
  }: AuthGuardAuth = children.type.auth;
  const isUser = !!session?.user;

  useEffect(() => {
    if (status === 'loading') return; // Do nothing while loading
    if (required && !isUser) unauthorized(); // If not authenticated, force log in
  }, [isUser, required, status, unauthorized]);

  if (!required || isUser) {
    return children;
  }

  // Session is being fetched, or no user.
  // If no user, useEffect() will redirect.
  return loading;
}
