import { useState, useEffect, ReactNode } from "react";
import { User, AuthState, LoginCredentials, RegisterData } from "@/types";
import { authApi } from "@/api/authApi";
import axiosInstance from "@/api/axios";
import { AuthContext } from "@/context/AuthContext";

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    token: localStorage.getItem("token"),
    isAuthenticated: !!localStorage.getItem("token"),
  });
  const [authLoaded, setAuthLoaded] = useState(false);

  // Sayfa yüklendiğinde localStorage'dan kullanıcı bilgilerini al
  useEffect(() => {
    const loadUser = async () => {
      const token = localStorage.getItem("token");
      const storedUser = localStorage.getItem("user");

      if (token && storedUser) {
        try {
          const user = JSON.parse(storedUser);
          setAuthState({
            token,
            user,
            isAuthenticated: true,
          });
        } catch (error) {
          console.error("Failed to parse stored user data:", error);
          localStorage.removeItem("user");
        }
      }

      // Yükleme tamamlandı
      setAuthLoaded(true);
    };

    loadUser();
  }, []);

  // Set authorization header whenever token changes
  useEffect(() => {
    if (authState.token) {
      axiosInstance.defaults.headers.common[
        "Authorization"
      ] = `Bearer ${authState.token}`;
    } else {
      delete axiosInstance.defaults.headers.common["Authorization"];
    }
  }, [authState.token]);

  const login = async (credentials: LoginCredentials) => {
    const response = await authApi.login(credentials);
    localStorage.setItem("token", response.token);
    localStorage.setItem("user", JSON.stringify(response.user)); // Kullanıcı bilgilerini kaydet
    setAuthState({
      token: response.token,
      user: response.user,
      isAuthenticated: true,
    });
  };

  const register = async (data: RegisterData) => {
    const response = await authApi.register(data);
    localStorage.setItem("token", response.token);
    localStorage.setItem("user", JSON.stringify(response.user)); // Kullanıcı bilgilerini kaydet
    setAuthState({
      token: response.token,
      user: response.user,
      isAuthenticated: true,
    });
  };

  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user"); // Kullanıcı bilgilerini temizle
    setAuthState({
      user: null,
      token: null,
      isAuthenticated: false,
    });
  };

  const updateUserProfile = (userData: Partial<User>) => {
    if (authState.user) {
      const updatedUser = { ...authState.user, ...userData };
      localStorage.setItem("user", JSON.stringify(updatedUser)); // Güncellenmiş kullanıcı bilgilerini kaydet
      setAuthState((prev) => ({
        ...prev,
        user: updatedUser,
      }));
    }
  };

  return (
    <AuthContext.Provider
      value={{
        ...authState,
        authLoaded,
        login,
        register,
        logout,
        updateUserProfile,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
