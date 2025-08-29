document.addEventListener("DOMContentLoaded", () => {
  const cardsContainer = document.getElementById("bandCards");
  const modal = document.getElementById("bandModal");
  const modalBody = document.getElementById("modalBody");
  const closeModal = document.querySelector(".modal-close");

  // Кэш для полученных данных о группах и концертах, чтобы уменьшить количество запросов
  let allBandsData = [];

  // --- Слушатели событий ---

  if (closeModal && modal) {
    // Закрыть модальное окно по клику на 'X'
    closeModal.addEventListener("click", () => hideBandModal());

    // Закрыть модальное окно по клику вне окна
    window.addEventListener("click", (e) => {
      if (e.target === modal) {
        hideBandModal();
      }
    });

    // Закрыть модальное окно по нажатию клавиши 'Escape'
    window.addEventListener("keydown", (e) => {
        if (e.key === "Escape" && modal.classList.contains("show-modal")) {
            hideBandModal();
        }
    });
  }

  // --- Загрузка страницы ---

  // Навешиваем обработчики на карточки, отрендеренные сервером
  document.querySelectorAll(".band-card-link").forEach((card) => {
    const bandID = parseInt(card.dataset.id, 10);
    card.addEventListener("click", (e) => {
      e.preventDefault();
      // Update URL and fetch data for the modal
      history.pushState({ bandId: bandID }, "", `/artist/${bandID}`);
      fetchAndShowBandModal(bandID);
    });
  });

  function renderArtists(artists) {
    if (!Array.isArray(artists)) return;
    cardsContainer.innerHTML = artists.map(artist => `
      <a class="band-card-link" href="/artist/${artist.id}" data-id="${artist.id}">
        <div class="band-card">
          <div class="band-header">
            <img src="${artist.image}" alt="${artist.name}" onerror="this.onerror=null;this.src='https://placehold.co/300x300/1f2937/e5e7eb?text=No+Photo';">
            <div class="band-name-overlay">
              <h2>${artist.name}</h2>
            </div>
          </div>
        </div>
      </a>
    `).join("");
    // Повторно навешиваем обработчики на новые карточки
    document.querySelectorAll(".band-card-link").forEach((card) => {
      const bandID = parseInt(card.dataset.id, 10);
      card.addEventListener("click", (e) => {
        e.preventDefault();
        history.pushState({ bandId: bandID }, "", `/artist/${bandID}`);
        fetchAndShowBandModal(bandID);
      });
    });
    // Обновляем счётчик
    const bandCount = document.getElementById("bandCount");
    if (bandCount) bandCount.textContent = artists.length;
  }

  function showError(msg) {
    cardsContainer.innerHTML = `<div class='error-container'><h2>Error</h2><p>${msg}</p></div>`;
    const bandCount = document.getElementById("bandCount");
    if (bandCount) bandCount.textContent = 0;
  }

  // --- Основные функции ---

  /**
   * Получает данные о конкретной группе по её ID и отображает модальное окно.
   * Кэширует данные, чтобы не запрашивать повторно.
   * @param {number} bandID - ID группы для отображения.
   */
  function fetchAndShowBandModal(bandID) {
    // Сначала пробуем найти группу в кэше
    const cachedBand = allBandsData.find((band) => band.id === bandID);
    if (cachedBand && cachedBand.concerts) {
      // Если группа и её концерты есть в кэше, показываем модалку сразу
      showBandModal(cachedBand);
      return;
    }

    // Если нет в кэше, запрашиваем детали группы
    fetch(`/filter?id=${bandID}`)
      .then((res) => {
        if (!res.ok) throw new Error(`Artist not found (HTTP ${res.status})`);
        return res.json();
      })
      .then((bands) => {
        if (!bands || bands.length === 0) throw new Error("Artist data is empty");
        const bandData = bands[0];
        // Кэшируем основные данные о группе
        if (!cachedBand) {
            allBandsData.push(bandData);
        }
        // Запрашиваем данные о концертах для этой группы
        return fetchConcerts(bandData);
      })
      .then((bandWithConcerts) => {
        // Показываем модалку со всеми данными
        showBandModal(bandWithConcerts);
      })
      .catch((err) => {
        console.error("Error fetching band details:", err);
        // Ошибка при получении данных о группе
        // Редирект на 404, если не удалось получить артиста (скорее всего, ID невалидный)
        window.location.href = '/404.html';
      });
  }

  /**
   * Получает данные о концертах для группы и добавляет их к объекту группы.
   * @param {object} bandData - объект группы.
   * @returns {Promise<object>} Промис с объектом группы и концертами.
   */
  function fetchConcerts(bandData) {
    return fetch(`/concerts/data?id=${bandData.id}`)
      .then((res) => {
        if (!res.ok) throw new Error(`Concerts not found (HTTP ${res.status})`);
        return res.json();
      })
      .then((concerts) => {
        // Добавляем концерты к объекту группы
        bandData.concerts = Array.isArray(concerts) ? concerts : [];
        return bandData;
      })
      .catch((err) => {
        console.error("Error fetching concerts:", err);
        // Если не удалось получить концерты, продолжаем без них
        bandData.concerts = [];
        return bandData;
      });
  }

  /**
   * Заполняет и отображает модальное окно с информацией о группе.
   * @param {object} band - объект группы с концертами.
   */
  function showBandModal(band) {
    const members = Array.isArray(band.members) ? band.members.join(", ") : "N/A";
    
    // Формируем HTML для временной шкалы концертов
    const concertsHTML = band.concerts && band.concerts.length > 0
      ? band.concerts.map(c => `
          <li>
            <span class="concert-date">${formatDate(c.date)}</span>
            <span class="concert-location">${formatLocation(c.location)}</span>
          </li>
        `).join("")
      : "<li>No upcoming concerts found.</li>";

    modalBody.innerHTML = `
      <img src="${band.image}" alt="${band.name}" onerror="this.onerror=null;this.src='https://placehold.co/250x250/1f2937/e5e7eb?text=Image+Not+Found';">
      <h2>${band.name}</h2>
      <p><strong>Formed:</strong> ${band.creationDate}</p>
      <p><strong>First Album:</strong> ${band.firstAlbum}</p>
      <p><strong>Members:</strong> ${members}</p>
      <h3>Concert History</h3>
      <ul id="concertList">${concertsHTML}</ul>
    `;
    modal.classList.add("show-modal");
  }

  /**
   * Скрывает модальное окно и сбрасывает URL.
   */
  function hideBandModal() {
    modal.classList.remove("show-modal");
    // Очищаем содержимое модального окна после анимации скрытия
    setTimeout(() => {
        modalBody.innerHTML = "";
    }, 300);
    // Сброс URL на главную страницу
    history.pushState({}, "", "/");
  }

  // --- Вспомогательные функции ---

  /**
   * Форматирует строку локации для отображения (например, "london-uk" -> "London, UK").
   * @param {string} loc - строка локации из API.
   * @returns {string} Отформатированная строка локации.
   */
  function formatLocation(loc) {
    if (!loc || typeof loc !== "string") return "Unknown Location";
    
    return loc
      .trim()
      .toLowerCase()
      .replace(/_/g, " ")
      .split('-') // Splits "city-country"
      .map(part => 
        part.split(" ")
            .map(word => {
                if (word === 'usa') return 'USA'; // Особый случай для USA
                return word.charAt(0).toUpperCase() + word.slice(1);
            })
            .join(" ")
      )
      .join(", "); // Joins parts with a comma
  }

  /**
   * Форматирует строку даты для отображения (например, "01-01-2023" -> "01 Jan 2023").
   * @param {string} dateStr - строка даты из API.
   * @returns {string} Отформатированная строка даты.
   */
  function formatDate(dateStr) {
    if (!dateStr || typeof dateStr !== "string") return "Unknown Date";
    const parts = dateStr.split("-");
    if (parts.length !== 3) return dateStr; // На случай неожиданного формата
    
    const [day, month, year] = parts;
    const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    const monthName = months[parseInt(month, 10) - 1] || "??";
    
    return `${day} ${monthName} ${year}`;
  }
});
