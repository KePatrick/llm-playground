
marked.setOptions({
	breaks: true,  // Enable line breaks
	gfm: true,     // Enable GitHub-flavored Markdown
	headerIds: false, // Avoid heading ID conflicts
	highlight: function(code, lang) {
		// If highlight.js is included
		if (lang && hljs.getLanguage(lang)) {
			return hljs.highlight(code, { language: lang }).value;
		}
		return code;
	}
});

function generateUUID() {
	// Generate a UUID v4
	return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
		var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
		return v.toString(16);
	});
}

// Get from sessionStorage or generate a new sessionId
function getSessionId() {
	let sessionId = window.sessionStorage.getItem('chatSessionId');
	if (!sessionId) {
		sessionId = generateUUID();
		window.sessionStorage.setItem('chatSessionId', sessionId);
	}
	return sessionId;
}

const chatContainer = document.getElementById('chat-container');
const messageInput = document.getElementById('message-input');
const sendButton = document.getElementById('send-button');
const themeToggle = document.getElementById('theme-toggle');


let currentResponseDiv = null;
let isProcessingStream = false;

const initialInputHeight = messageInput.scrollHeight + 'px';
messageInput.style.height = 'auto';
messageInput.style.height = initialInputHeight;

const maxInputHeight = 132
messageInput.addEventListener('input', () => {
	messageInput.style.height = 'auto';
	if(messageInput.scrollHeight > maxInputHeight) {
		messageInput.style.height = maxInputHeight + 'px'
	}else{

		messageInput.style.height = messageInput.scrollHeight + 'px';
	}
});

async function sendMessage() {
	
	const prompt = messageInput.value.replace(/\r\n/g, '\n');
	// const prompt = messageInput.value.trim();
	if (!prompt || isProcessingStream) return;

	const sessionId = getSessionId();
	addMessage(prompt, 'user-message');
	messageInput.value = '';

	messageInput.style.height = 'auto';
	messageInput.style.height = initialInputHeight;

	isProcessingStream = true;
	sendButton.disabled = true;
	messageInput.disabled = true;

	currentResponseDiv = document.createElement('div');
	currentResponseDiv.classList.add('message', 'bot-message');
	chatContainer.appendChild(currentResponseDiv);

	const cursor = document.createElement('span');
	cursor.classList.add('cursor');
	currentResponseDiv.appendChild(cursor);

	const requestData = { prompt: prompt , sessionId: sessionId};

	try {
		const response = await fetch(apiUrl, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(requestData),
		});

		if (!response.ok) {
			throw new Error(`API request failed: ${response.status}`);
		}

		const reader = response.body.getReader();
		const decoder = new TextDecoder();

		let markdownContent = '';

		readStream:
		while (true) {
			const { value, done } = await reader.read();
			if (done) break;
			const chunk = decoder.decode(value, { stream: true });
			console.log('Raw chunk:', JSON.stringify(chunk));
			const lines = chunk.split('\n\n').filter(Boolean);

			for (const line of lines) {
				if (line.startsWith('data: ')) {
					const data = line.substring(6);
					console.log('Received:', JSON.stringify(data));

					if (data === '[DONE]') {
						break readStream;
					}

					// Replace newline placeholders
					const restoredData = data.replace(/\[NEWLINE\]/g, '\n');
					markdownContent += restoredData;

					// Parse Markdown
					const htmlContent = marked.parse(markdownContent);

					currentResponseDiv.innerHTML = htmlContent;
					currentResponseDiv.appendChild(cursor);

					document.querySelectorAll('pre code').forEach((block) => {
						hljs.highlightBlock(block);
					});

					chatContainer.scrollTop = chatContainer.scrollHeight;
				}
			}
		}

		if (cursor.parentNode === currentResponseDiv) {
			currentResponseDiv.removeChild(cursor);
		}

	} catch (error) {
		currentResponseDiv.textContent = 'error: ' + error.message;
	} finally {
		isProcessingStream = false;
		sendButton.disabled = false;
		messageInput.disabled = false;
		messageInput.focus();
		chatContainer.scrollTop = chatContainer.scrollHeight;
	}



}

function addMessage(text, className) {
	const messageDiv = document.createElement('div');
	messageDiv.classList.add('message', className);
	messageDiv.innerHTML = text.replace(/\n/g, '<br>');
	chatContainer.appendChild(messageDiv);

	if (className === 'user-message') {
		requestAnimationFrame(() => {
			messageDiv.scrollIntoView({
				behavior: 'smooth',
				block: 'start'
			});
		});	
	}
	return messageDiv;
}

sendButton.addEventListener('click', sendMessage);
messageInput.addEventListener('keydown', (event) => {
	if (event.key === 'Enter' && !event.shiftKey) {
		sendMessage();
	}
});

// Theme switcher
themeToggle.addEventListener('click', () => {
	const currentTheme = document.body.getAttribute('data-theme');
	if (currentTheme === 'light') {
		document.body.removeAttribute('data-theme');
		localStorage.setItem('theme', 'dark');
	} else {
		document.body.setAttribute('data-theme', 'light');
		localStorage.setItem('theme', 'light');
	}
});

// Apply saved theme on load
const savedTheme = localStorage.getItem('theme');
if (savedTheme === 'light') {
	document.body.setAttribute('data-theme', 'light');
}

