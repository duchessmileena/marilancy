
async function register() {
  alert("REGISTER DIPANGGIL");

  const namaEl = document.getElementById("nama");
  const emailEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");
  const roleEl = document.getElementById("role");

  console.log("ELEMENT:", {
    nama: namaEl,
    email: emailEl,
    password: passwordEl,
    role: roleEl
  });

  if (!namaEl || !emailEl || !passwordEl || !roleEl) {
    alert("ADA ELEMENT YANG TIDAK TERBACA!");
    return;
  }

  const data = {
    nama: namaEl.value,
    email: emailEl.value,
    password: passwordEl.value,
    role: roleEl.value,
  };

  console.log("DATA DIKIRIM:", data);
  alert("ROLE DIPILIH: " + data.role);

  try {
    const res = await fetch("/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });

    const result = await res.json();

    console.log("RESPONSE:", result);

    if (result.error) {
      alert("ERROR: " + result.error);
    } else {
      alert(result.message);
      window.location = "/";
    }

  } catch (err) {
    console.error("FETCH ERROR:", err);
    alert("Gagal koneksi ke server");
  }
}

async function login() {
  const emailEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");
  const roleEl = document.getElementById("role");

  if (!emailEl || !passwordEl || !roleEl) {
    alert("Element login tidak terbaca");
    return;
  }

  const data = {
    email: emailEl.value,
    password: passwordEl.value,
    role: roleEl.value,
  };

  console.log("LOGIN DATA:", data);

  try {
    const res = await fetch("/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });

    const result = await res.json();

    console.log("LOGIN RESPONSE:", result);

    if (result.role === "freelancer") window.location = "/freelancer";
    else if (result.role === "client") window.location = "/client";
    else if (result.role === "admin") window.location = "/admin";
    else alert(result.error);

  } catch (err) {
    console.error("LOGIN ERROR:", err);
    alert("Gagal login");
  }
}