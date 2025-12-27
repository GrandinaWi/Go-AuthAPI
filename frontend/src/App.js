import { useState, useEffect } from "react";

const API_URL = "http://localhost:8080";

export default function App() {
  const [mode, setMode] = useState("login"); // login | register
  const [form, setForm] = useState({
    username: "",
    password: "",
    age: "",
  });
  const [token, setToken] = useState(null);
  const [user, setUser] = useState(null);
  const [error, setError] = useState("");

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const submit = async () => {
    setError("");

    const res = await fetch(`${API_URL}/${mode}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: form.username,
        password: form.password,
        age: Number(form.age),
      }),
    });

    if (!res.ok) {
      setError("Ошибка авторизации");
      return;
    }

    const data = await res.json();
    setToken(data.token);
  };

  useEffect(() => {
    if (!token) return;

    fetch(`${API_URL}/user`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
        .then((res) => res.json())
        .then(setUser)
        .catch(() => setError("Ошибка загрузки пользователя"));
  }, [token]);

  return (
      <div style={styles.page}>
        <div style={styles.card}>
          <h2>{user ? "Профиль" : mode === "login" ? "Вход" : "Регистрация"}</h2>

          {!user && (
              <>
                <input
                    style={styles.input}
                    name="username"
                    placeholder="Логин"
                    onChange={handleChange}
                />
                <input
                    style={styles.input}
                    name="password"
                    type="password"
                    placeholder="Пароль"
                    onChange={handleChange}
                />
                {mode === "register" && (
                    <input
                        style={styles.input}
                        name="age"
                        placeholder="Возраст"
                        onChange={handleChange}
                    />
                )}

                <button style={styles.button} onClick={submit}>
                  {mode === "login" ? "Войти" : "Зарегистрироваться"}
                </button>

                <p
                    style={styles.switch}
                    onClick={() =>
                        setMode(mode === "login" ? "register" : "login")
                    }
                >
                  {mode === "login"
                      ? "Нет аккаунта? Регистрация"
                      : "Уже есть аккаунт? Войти"}
                </p>
              </>
          )}

          {user && (
              <div style={styles.profile}>
                <p>
                  <b>Логин:</b> {user.username}
                </p>
                <p>
                  <b>Возраст:</b> {user.age}
                </p>
                <button
                    style={styles.logout}
                    onClick={() => {
                      setUser(null);
                      setToken(null);
                    }}
                >
                  Выйти
                </button>
              </div>
          )}

          {error && <p style={styles.error}>{error}</p>}
        </div>
      </div>
  );
}

const styles = {
  page: {
    height: "100vh",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    background: "#f5f6fa",
    fontFamily: "sans-serif",
  },
  card: {
    width: 320,
    padding: 20,
    borderRadius: 10,
    background: "#fff",
    boxShadow: "0 10px 30px rgba(0,0,0,0.1)",
  },
  input: {
    width: "100%",
    padding: 10,
    marginBottom: 10,
    borderRadius: 5,
    border: "1px solid #ddd",
  },
  button: {
    width: "100%",
    padding: 10,
    borderRadius: 5,
    background: "#4f46e5",
    color: "#fff",
    border: "none",
    cursor: "pointer",
  },
  switch: {
    marginTop: 10,
    fontSize: 14,
    color: "#4f46e5",
    cursor: "pointer",
    textAlign: "center",
  },
  profile: {
    marginTop: 10,
  },
  logout: {
    marginTop: 10,
    width: "100%",
    padding: 8,
    border: "none",
    borderRadius: 5,
    background: "#e11d48",
    color: "#fff",
    cursor: "pointer",
  },
  error: {
    marginTop: 10,
    color: "red",
    fontSize: 14,
  },
};
