import React, { createContext, useContext, useEffect, useState } from 'react';
import Backend from '../app/clients/backend/backend';
import { User } from '../app/clients/backend/types';

interface AuthContextType {
    user: User | null;
    loading: boolean;
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

    const refreshUser = async () => {
        try {
            const response = await backend.getCurrentUser();
            setUser(response.user);
        } catch (error) {
            setUser(null);
        }
    };

    useEffect(() => {
        const checkAuth = async () => {
            try {
                await refreshUser();
            } catch (error) {
                console.error('Auth check failed:', error);
            } finally {
                setLoading(false);
            }
        };

        checkAuth();
    }, []);

    const login = async (email: string, password: string) => {
        // Backend sets JWT token in HTTP-only cookie (valid for 30 days)
        const response = await backend.login({ email, password });
        setUser(response.user);
    };

    const register = async (email: string, password: string) => {
        // Backend sets JWT token in HTTP-only cookie (valid for 30 days)
        const response = await backend.register({ email, password });
        setUser(response.user);
    };

    const logout = async () => {
        // Backend clears JWT token cookie
        await backend.logout();
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, loading, login, register, logout, refreshUser }}>
            {children}
        </AuthContext.Provider>
    );
};

