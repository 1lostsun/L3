const API_BASE = "http://localhost:8080/tree_comments/api"; // твой backend

let currentPage = 1;
const pageSize = 5;
let currentSort = "asc";
let currentSearch = "";

document.addEventListener("DOMContentLoaded", () => {
    loadComments();

    document.getElementById("searchBtn").onclick = () => {
        currentSearch = document.getElementById("searchInput").value.trim();
        currentPage = 1;
        loadComments();
    };

    document.getElementById("sortSelect").onchange = e => {
        currentSort = e.target.value;
        loadComments();
    };

    document.getElementById("prevPage").onclick = () => {
        if (currentPage > 1) {
            currentPage--;
            loadComments();
        }
    };

    document.getElementById("nextPage").onclick = () => {
        currentPage++;
        loadComments();
    };

    document.getElementById("addCommentBtn").onclick = () => addComment();
});

async function loadComments(parent = null) {
    const params = new URLSearchParams({
        page: currentPage,
        limit: pageSize,
        sort: currentSort,
        search: currentSearch,
    });

    if (parent) params.append("parent", parent);

    const response = await fetch(`${API_BASE}/comments?page=${currentPage}&limit=${pageSize}&sort=${currentSort}&search=${currentSearch}`);
    const data = await response.json();
    console.log(response)

    const comments = Array.isArray(data.comments) ? data.comments : [];
    const commentsContainer = document.getElementById("commentsContainer");

    renderComments(comments, commentsContainer);
}

function renderComments(comments, container, level = 0) {
    container.innerHTML = "";
    comments.forEach(comment => {
        const div = document.createElement("div");
        div.className = "comment";
        div.style.marginLeft = `${level * 20}px`;

        div.innerHTML = `
      <div class="meta">#${comment.id} • ${new Date(comment.date).toLocaleString()}</div>
      <div class="text">${comment.text}</div>
      <div class="actions">
        <button onclick="showReplyBox('${comment.id}')">Ответить</button>
        <button onclick="deleteComment('${comment.id}')">Удалить</button>
      </div>
      <div id="reply-${comment.id}" class="reply-box" style="display:none">
        <textarea id="replyText-${comment.id}" placeholder="Ответ..."></textarea>
        <button onclick="sendReply('${comment.id}')">Отправить</button>
      </div>
    `;

        container.appendChild(div);

        if (comment.commentsTree?.length) {
            const childContainer = document.createElement("div");
            childContainer.className = "reply";
            renderComments(comment.commentsTree, childContainer, level + 1);
            container.appendChild(childContainer);
        }
    });

    document.getElementById("pageInfo").textContent = `Страница ${currentPage}`;
}

function showReplyBox(id) {
    const box = document.getElementById(`reply-${id}`);
    box.style.display = box.style.display === "none" ? "block" : "none";
}

async function sendReply(parentId) {
    const text = document.getElementById(`replyText-${parentId}`).value.trim();
    if (!text) return alert("Введите текст!");

    await fetch(`${API_BASE}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ parent: parentId, text }),
    });

    loadComments();
}

async function addComment() {
    const text = document.getElementById("newCommentText").value.trim();
    if (!text) return alert("Введите текст!");

    await fetch(`${API_BASE}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text }),
    });

    document.getElementById("newCommentText").value = "";
    loadComments();
}

async function deleteComment(id) {
    if (!confirm("Удалить комментарий и все ответы?")) return;

    await fetch(`${API_BASE}/comments/${id}`, { method: "DELETE" });
    loadComments();
}
