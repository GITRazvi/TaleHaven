import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import Backend from '../app/clients/backend/backend';
import { User } from '../app/clients/backend/types';
import { sessionExpiredEvent } from './auth-session-events';

interface AuthContextType {
    user: User | null;
    loading: boolean;
    sessionExpired: boolean;
    login: (email: string, password: string) => Promise<void>;
    register: (email: string, password: string) => Promise<void>;
    logout: () => Promise<void>;
    refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

const backend = new Backend();

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [sessionExpired, setSessionExpired] = useState(false);

    const refreshUser = useCallback(async () => {
        try {
            const response = await backend.getCurrentUser();
            setUser(response.user);
            setSessionExpired(false);
        } catch (error) {
            setUser(null);
        }
    }, []);

    useEffect(() => {
        const handleSessionExpired = () => {
            setUser(null);
            setSessionExpired(true);
        };

        window.addEventListener(sessionExpiredEvent, handleSessionExpired);
        refreshUser().finally(() => setLoading(false));

        return () => window.removeEventListener(sessionExpiredEvent, handleSessionExpired);
    }, [refreshUser]);

    const login = async (email: string, password: string) => {
        const response = await backend.login({ email, password });
        setUser(response.user);
        setSessionExpired(false);
    };

    const register = async (email: string, password: string) => {
        const response = await backend.register({ email, password });
        setUser(response.user);
        setSessionExpired(false);
    };

    const logout = async () => {
        try {
            await backend.logout();
        } finally {
            setUser(null);
            setSessionExpired(false);
        }
    };

    return (
        <AuthContext.Provider value={{ user, loading, sessionExpired, login, register, logout, refreshUser }}>
            {children}
        </AuthContext.Provider>
    );
};

