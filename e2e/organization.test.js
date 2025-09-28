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

const upsertOrganization = async (token) => {
  const response = await axios.post(
    `${BASE_URL}/organization/upsert`,
    {
      name: "Test Organization",
      description: "Test Description",
      client_id: "test",
      tenant_id: "test",
      client_secret: "test",
    },
    {
      headers: {
        Authorization: `${token}`,
      },
    }
  );
  return response.data.id;
};
describe("Organization module tests", () => {
  it("should upsert an organization", async () => {
    const token = await getToken();
    const response = await axios.post(
      `${BASE_URL}/organization/upsert`,
      {
        name: "Test Organization",
        description: "Test Description",
        client_id: CLIENT_ID,
        tenant_id: TENANT_ID,
        client_secret: CLIENT_SECRET,
      },
      {
        headers: {
          Authorization: `${token}`,
        },
      }
    );
    expect(response.status).toBe(200);
    expect(response.data.id).toBeDefined();
  });

  it("should get an organization", async () => {
    const token = await getToken();
    const id = await upsertOrganization(token);
    const response = await axios.get(`${BASE_URL}/organization/get`, {
      headers: {
        Authorization: `${token}`,
      },
    });
    expect(response.status).toBe(200);
    expect(response.data.id).toBe(id);
    expect(response.data.name).toBe("Test Organization");
    expect(response.data.description).toBe("Test Description");
    expect(response.data.client_id).toBeDefined();
    expect(response.data.tenant_id).toBeDefined();
    expect(response.data.is_authorized).toBe(false);
  });

  it("should delete an organization", async () => {
    const token = await getToken();
    const response = await axios.delete(`${BASE_URL}/organization/delete`, {
      headers: {
        Authorization: `${token}`,
      },
    });
  });
});
