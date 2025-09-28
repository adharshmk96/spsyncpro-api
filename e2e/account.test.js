const axios = require("axios");
const BASE_URL = "http://localhost:8080/api/v1";

const getToken = async () => {
  const email = `test${Math.random()}@test.com`;
  const response = await axios.post(`${BASE_URL}/account/register`, {
    email: email,
    password: "password",
  });
  return response.data.token;
};

describe("Account module tests", () => {
  it("should register a new account", async () => {
    const email = `test${Math.random()}@test.com`;
    const response = await axios.post(`${BASE_URL}/account/register`, {
      email: email,
      password: "password",
    });
    expect(response.status).toBe(200);
    expect(response.data.email).toBe(email);
    expect(response.data.id).toBeDefined();
    expect(response.data.token).toBeDefined();
  });

  it("should login a new account", async () => {
    await axios.post(`${BASE_URL}/account/register`, {
      email: "test@test.com",
      password: "password",
    });
    const response = await axios.post(`${BASE_URL}/account/login`, {
      email: "test@test.com",
      password: "password",
    });
    expect(response.status).toBe(200);
    expect(response.data.token).toBeDefined();
  });

  it("should get profile of a new account", async () => {
    const token = await getToken();
    const response = await axios.get(`${BASE_URL}/account/profile`, {
      headers: {
        Authorization: `${token}`,
      },
    });
    expect(response.status).toBe(200);
    expect(response.data.email).toBeDefined();
    expect(response.data.id).toBeDefined();
  });

  it("should logout a new account", async () => {
    const token = await getToken();
    console.log(" token", token);
    const response = await axios.post(
      `${BASE_URL}/account/logout`,
      {},
      {
        headers: {
          Authorization: `${token}`,
        },
      }
    );
    expect(response.status).toBe(200);
    expect(response.data.message).toBe("logout successful");
  });

  it("should change password of a new account", async () => {
    const token = await getToken();
    const response = await axios.post(
      `${BASE_URL}/account/change-password`,
      {
        old_password: "password",
        new_password: "newpassword",
      },
      {
        headers: {
          Authorization: `${token}`,
        },
      }
    );
    expect(response.status).toBe(200);
    expect(response.data.message).toBe("password changed successfully");
  });
});
