// Frontend calls are routed via nginx to backend: /api -> backend:8080/img_compressor/v1
(function(){
    const form = document.getElementById('uploadForm');
    const statusSection = document.getElementById('statusSection');
    const resultSection = document.getElementById('resultSection');
    const resultImg = document.getElementById('resultImg');
    const deleteBtn = document.getElementById('deleteBtn');

    let currentImageId = null;
    let pollTimer = null;

    function setStatus(text){
        statusSection.textContent = text;
        statusSection.classList.remove('hidden');
    }

    function clearStatus(){
        statusSection.classList.add('hidden');
        statusSection.textContent = '';
    }

    function showResult(imageId){
        const url = `/api/image/${encodeURIComponent(imageId)}`;
        resultImg.src = url;
        resultSection.classList.remove('hidden');
    }

    async function upload(formData){
        const res = await fetch('/api/upload', { method: 'POST', body: formData });
        if(!res.ok){
            const data = await res.json().catch(()=>({error: res.statusText}));
            throw new Error(data.error || 'Upload failed');
        }
        return res.json();
    }

    async function checkStatus(imageId){
        const res = await fetch(`/api/image/${encodeURIComponent(imageId)}`);
        if(!res.ok){
            const data = await res.json().catch(()=>({error: res.statusText}));
            throw new Error(data.error || 'Status failed');
        }
        const contentType = res.headers.get('content-type') || '';
        if(contentType.includes('application/json')){
            return res.json();
        }
        // If not JSON, backend returned the file directly (done)
        return 'done-file';
    }

    function startPolling(imageId){
        if(pollTimer) clearInterval(pollTimer);
        pollTimer = setInterval(async ()=>{
            try{
                const data = await checkStatus(imageId);
                if(data === 'done-file'){
                    clearInterval(pollTimer);
                    clearStatus();
                    showResult(imageId);
                    return;
                }
                const status = data.status || 'unknown';
                setStatus(`Статус: ${status}`);
                if(status === 'done'){
                    clearInterval(pollTimer);
                    clearStatus();
                    showResult(imageId);
                }
                if(status === 'failed'){
                    clearInterval(pollTimer);
                    setStatus('Обработка не удалась');
                }
            }catch(err){
                clearInterval(pollTimer);
                setStatus(err.message);
            }
        }, 1500);
    }

    form.addEventListener('submit', async (e)=>{
        e.preventDefault();
        clearStatus();
        resultSection.classList.add('hidden');
        currentImageId = null;

        const file = document.getElementById('file').files[0];
        const width = parseInt(document.getElementById('width').value, 10);
        const height = parseInt(document.getElementById('height').value, 10);
        const watermark = document.getElementById('watermark').value.trim();

        if(!file){
            setStatus('Выберите файл');
            return;
        }

        const operations = {};
        if(Number.isFinite(width) || Number.isFinite(height)){
            operations.resize = { width: Number.isFinite(width) ? width : 0, height: Number.isFinite(height) ? height : 0 };
        }
        if(watermark){
            operations.watermark = { text: watermark };
        }
        if(!operations.resize && !operations.watermark){
            setStatus('Укажите хотя бы одну операцию');
            return;
        }

        const fd = new FormData();
        fd.append('file', file);
        fd.append('operations', JSON.stringify(operations));

        try{
            setStatus('Загрузка...');
            const data = await upload(fd);
            currentImageId = data.image_id;
            setStatus(`Создана задача: ${currentImageId}`);
            startPolling(currentImageId);
        }catch(err){
            setStatus(err.message);
        }
    });

    deleteBtn.addEventListener('click', async ()=>{
        if(!currentImageId) return;
        try{
            const res = await fetch(`/api/image/${encodeURIComponent(currentImageId)}`, { method: 'DELETE' });
            if(!res.ok){
                const data = await res.json().catch(()=>({error: res.statusText}));
                throw new Error(data.error || 'Delete failed');
            }
            setStatus('Удалено');
            resultSection.classList.add('hidden');
            currentImageId = null;
        }catch(err){
            setStatus(err.message);
        }
    });
})();


