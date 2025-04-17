// services/auth.js
import axios from "axios";

const API_URL = "http://localhost:8080/api";

interface LoginResponse {
  token: string;
  user: User;
}

interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
}

interface RegisterResponse {
  message: string;
  user: User;
}

export const login = async (
  username: string,
  password: string
): Promise<LoginResponse> => {
  const response = await axios.post<LoginResponse>(`${API_URL}/users/login`, {
    username,
    password,
  });
  if (response.data.token) {
    localStorage.setItem("token", response.data.token);
    localStorage.setItem("user", JSON.stringify(response.data.user));
  }
  return response.data;
};

export const register = async (
  username: string,
  email: string,
  password: string,
  fullName: string
): Promise<RegisterResponse> => {
  const response = await axios.post<RegisterResponse>(
    `${API_URL}/users/register`,
    {
      username,
      email,
      password,
      full_name: fullName,
    }
  );
  return response.data;
};

export const getCurrentUser = () => {
  const userStr = localStorage.getItem("user");
  return userStr ? JSON.parse(userStr) : null;
};

export const logout = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("user");
};
