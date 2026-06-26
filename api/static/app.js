let token = '';
let currentPath = '/';
let currentUser = null; // 存储当前登录用户信息
let deptList = []; // 存储部门列表
let empList = []; // 存储员工列表

// 更智能的 token 续期管理
let tokenRefreshTimer = null;

// Media Player functionality
let playlist = [];
let currentTrackIndex = -1;
let isRandomPlay = false;

function setupTokenRefresh(expiresIn) {
    if (tokenRefreshTimer) {
        clearInterval(tokenRefreshTimer);
    }
    
    // 在过期前1分钟刷新
    const refreshTime = (expiresIn - 60) * 1000;
    if (refreshTime > 0) {
        tokenRefreshTimer = setTimeout(keepLogin, refreshTime);
    } else {
        // token 已经过期或即将过期，立即刷新
        keepLogin();
    }
}

// 获取验证码
function getCaptcha(type) {
    $.get('/api/v1/user/captcha', function(response) {
        console.log(response);
        if (response.code === 200) {
            // console.log(response.data.imageBase64);
            if (type === 'login') {
                $('#loginCaptchaId').val(response.data.captchaId);
                $('#captchaImage').attr('src',  response.data.imageBase64);
                $('#loginCaptcha').val("");

            } else {
                $('#regCaptchaId').val(response.data.captchaId);
                $('#regCaptchaImage').attr('src', response.data.imageBase64);
                $('#regCaptcha').val("");
            }
        }
    });
}

// 添加 token 续期函数
function keepLogin() {
    $.ajax({
        url: '/api/v1/user/keeplogin',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        success: function(response) {
            responseStr = JSON.stringify(response);
            if (response.code === 200) {
                // 更新 token
                token = response.data.accessToken;
                localStorage.setItem('token', token);
            } else if (response.code === 401)  {
                // token 已失效，需要重新登录
                alert('登录已过期，请重新登录 keeplogin');
                logout();
            } else {
                alert('登录已过期，请重新登录 keeplogin2'+response.code);
                logout();
            }
        },
        error: function() {
            // 请求失败，可能需要重新登录
            logout();
        }
    });
}

// 登录
function login() {
    const loginData = {
        empCode: $('#loginEmail').val(),
        password: $('#loginPassword').val(),
        captcha: $('#loginCaptcha').val(),
        captchaId: $('#loginCaptchaId').val()
    };

    $.ajax({
        url: '/api/knowsource/login',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(loginData),
        success: function(response) {
            if (response.code === 200) {
                token = response.data.token;
                currentUser = response.data.userInfo;
                localStorage.setItem('token', token);
                localStorage.setItem('userInfo', JSON.stringify(currentUser));
                $('#loginForm').hide();
                $('#fileManager').show();
                
                // 显示用户信息
                updateUserInfo();

                // 检查 URL 中是否有 path 参数
                const urlParams = new URLSearchParams(window.location.search);
                const pathParam = urlParams.get('path');
                if (pathParam) {
                    currentPath = decodeURIComponent(pathParam);
                }

                listFiles(currentPath);

                // 登录成功后获取部门和员工列表
                loadDeptList();
                loadEmpList();

                // 设置定时刷新 token（每9分钟刷新一次，假设token有效期为10分钟）
                startKeepLoginInterval();
            } else {
                alert('登录失败: ' + response.message);
                getCaptcha('login');
            }
        }
    });
}

// 注册
function register() {
    const registerData = {
        username: $('#regUsername').val(),
        email: $('#regEmail').val(),
        password: $('#regPassword').val(),
        captcha: $('#regCaptcha').val(),
        captchaId: $('#regCaptchaId').val()
    };

    $.ajax({
        url: '/api/v1/user/register',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(registerData),
        success: function(response) {
            responseStr = JSON.stringify(response);
            if (response.code === 200) {
                alert('注册成功，请登录');
                $('#registerForm').hide();
                $('#loginForm').show();
                getCaptcha('login');
            } else {
                console.log(responseStr);
                alert('注册失败: ' + response.message);
                getCaptcha('register');
            }
        }
    });
}

// 登出
function logout() {
    token = '';
    currentUser = null;
    // 清除 localStorage 中的 token 和用户信息
    localStorage.removeItem('token');
    localStorage.removeItem('userInfo');
    $('#fileManager').hide();
    $('#loginForm').show();
    getCaptcha('login');
}

// 更新用户信息显示
function updateUserInfo() {
    if (currentUser) {
        const userInfoText = `${currentUser.empName} (${currentUser.empCode})`;
        $('#userInfo').text(userInfoText);
    } else {
        $('#userInfo').text('');
    }
}

// 获取部门列表
function loadDeptList() {
    $.ajax({
        url: '/api/knowsource/dept/list',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            page: 1,
            pageSize: 1000  // 获取所有部门
        }),
        success: function(response) {
            if (response.code === 200 && response.data) {
                deptList = response.data.list || [];
                console.log('部门列表加载成功，共', deptList.length, '个部门');
                // 可以在这里处理部门列表数据
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录');
                logout();
            } else {
                console.error('获取部门列表失败:', response.message);
            }
        },
        error: function(xhr, status, error) {
            console.error('获取部门列表错误:', error);
        }
    });
}

// 获取员工列表
function loadEmpList() {
    $.ajax({
        url: '/api/knowsource/emp/list',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            page: 1,
            pageSize: 1000  // 获取所有员工
        }),
        success: function(response) {
            if (response.code === 200 && response.data) {
                empList = response.data.list || [];
                console.log('员工列表加载成功，共', empList.length, '个员工');
                // 可以在这里处理员工列表数据
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录');
                logout();
            } else {
                console.error('获取员工列表失败:', response.message);
            }
        },
        error: function(xhr, status, error) {
            console.error('获取员工列表错误:', error);
        }
    });
}

function displayCurrentPath(path) {
    const parts = path.split('/').filter(p => p);
    let html = '<a href="#" onclick="navigateDirectory(\'/\')">根目录</a>';
    let currentPath = '';
    
    parts.forEach(part => {
        currentPath += '/' + part;
        html += ' / <a href="#" onclick="navigateDirectory(\'' + currentPath + '\')">' + part + '</a>';
    });
    
    $('#currentPath').html('' + html);
}


// 修改 listFiles 函数中显示路径的部分
function listFiles(path) {
    $.ajax({
        url: '/api/v1/directory/list',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            path: path,
            page: 1,
            pageSize: 50
        }),
        success: function(response) {
            if (response.code === 200) {
                displayFiles(response.data.files);
                currentPath = response.data.currentPath;
                displayCurrentPath(currentPath); // 使用新的函数替换原来的文本显示
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录 listFiles');
                logout();
            } else {
                alert('登录已过期，请重新登录 listFiles2'+response);
                logout();
            }
        }
    });
}

// 搜索文件
function searchFiles(query) {
    // 如果没有传入 query 参数，则从搜索框获取
    if (!query) {
        query = $('#searchInput').val();
    }
    
    if (!query) {
        listFiles(currentPath);
        // 清除 URL 中的搜索参数
        history.pushState(null, '', window.location.pathname);
        return;
    }

    // 更新 URL
    history.pushState(null, '', `${window.location.pathname}?search=${encodeURIComponent(query)}`);

    $.ajax({
        url: '/api/v1/files/search',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            query: query,
            path: "/",
            page: 1,
            pageSize: 50
        }),
        success: function(response) {

            //把respone 转换为字符串
            var responseStr = JSON.stringify(response);


            if (response.code === 200) {
                displayFiles(response.data.files);
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录 search' );
                logout();
            } else {
                alert('登录已过期，请重新登录 search2'+responseStr);
                logout();
                //alert('获取文件列表失败: ' + response.code + " " + response.message);
            }
        }
    });
}

// 显示文件列表
function displayFiles(files) {
    const fileList = $('#fileList');
    fileList.empty();

    // 添加返回上级目录的选项（如果不在根目录）
    if (currentPath !== '/') {
        fileList.append(`
            <div class="file-item directory" onclick="navigateDirectory('..')">
                <div class="file-info">
                    <span>..</span>
                </div>
            </div>
        `);
    }

    files.forEach(file => {

        const escapedPath = file.path.replace(/'/g, "\\'");
        // console.log("escapedPath", escapedPath);
        
        const fileItem = $(`
            <div class="file-item ${file.isDirectory ? 'directory' : 'file'}" style="display: flex; justify-content: space-between; align-items: center;">
                <div class="file-info" onclick="${file.isDirectory ? 'navigateDirectory' : 'handleFile'}('${file.path}')">
                    <span>${file.name}</span>
                    ${!file.isDirectory ? `<span class="file-size">(${formatSize(file.size)})</span>` : ''}
                </div>
                <div class="file-actions">
                    ${!file.isDirectory ? `
                        <div class="dropdown">
                            <button class="btn btn-sm btn-primary dropdown-toggle like-btn" type="button" data-bs-toggle="dropdown" aria-expanded="false">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                                </svg>
                                <span class="like-label">喜欢</span>
                            </button>
                            <ul class="dropdown-menu">
                                <li><a class="dropdown-item" href="#" onclick="likeFile('${file.path}', 5)">Obsessed (超级着迷的)</a></li>
                                <li><a class="dropdown-item" href="#" onclick="likeFile('${file.path}', 4)">Awesome (超棒的)</a></li>
                                <li><a class="dropdown-item" href="#" onclick="likeFile('${file.path}', 3)">Enjoyable (令人享受的)</a></li>
                                <li><a class="dropdown-item" href="#" onclick="likeFile('${file.path}', 2)">Catchy (朗朗上口，有吸引力)</a></li>
                                <li><a class="dropdown-item" href="#" onclick="likeFile('${file.path}', 1)">Nice (还不错)</a></li>
                            </ul>
                        </div>
                        <button class="btn btn-sm btn-primary" 
                                onclick="navigateDirectory('${file.path.substring(0, file.path.lastIndexOf('/'))}')">
                            目录
                        </button>
                    ` : ''}
                </div>
            </div>
        `);
        fileList.append(fileItem);
    });
}

// 导航到目录
function navigateDirectory(path) {
    if (path === '..') {
        const parts = currentPath.split('/');
        parts.pop();
        path = parts.join('/') || '/';
    }

    // 更新 URL，添加 path 参数
    history.pushState(null, '', `${window.location.pathname}?path=${encodeURIComponent(path)}`);

    $.ajax({
        url: '/api/v1/directory/list',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            path: path
        }),
        success: function(response) {
            if (response.code === 200) {
                listFiles(path);
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录');
                logout();
            } else {
                alert('登录已过期，请重新登录 navigateDirectory'+response);
                logout();
                //alert('获取文件列表失败: ' + response.code + " " + response.message);
            }
        }
    });
}

const isWeiXin = function() {
    var ua = window.navigator.userAgent.toLowerCase();
    if (ua.match(/MicroMessenger/i) == 'micromessenger') {
        return true;
    } else {
        return false;
    }
}

// 处理文件（可以是播放或其他操作）
async function handleFile(path) {
    console.log("handleFile", path);
    const isMediaFile = path.toLowerCase().match(/\.(m4a|mp3|mp4|flac)$/);
    
    if (isMediaFile) {
        // Find all media files in the current directory
        const mediaFiles = $('#fileList .file-item').not('.directory').filter(function() {
            return $(this).find('.file-info span:first').text().toLowerCase().match(/\.(m4a|mp3|mp4|flac)$/);
        }).map(function() {
            return {
                name: $(this).find('.file-info span:first').text(),
                path: $(this).find('.file-info').attr('onclick').match(/'([^']+)'/)[1]
            };
        }).get();

        // Update playlist and play selected file
        playlist = mediaFiles;
        currentTrackIndex = playlist.findIndex(file => file.path === path);
        playTrack(currentTrackIndex);
    } else {
        file_ext = path.split('.').pop();

        // Handle non-media files as before
        $.ajax({
            url: '/api/v1/files/apply',
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token
            },
            contentType: 'application/json',
            data: JSON.stringify({
                file: path
            }),
            success: function(response) {
                if (response.code != "") {
                    const newurl = `/api/v1/filesget/${encodeURIComponent(path)}?code=${response.code}&ext=.${file_ext}`;
                    if (isWeiXin()) {
                        window.location.href = newurl;
                    } else {
                        window.open(newurl, 'listenbook');
                    }
                } else {
                    alert('获取文件访问码失败: ' + response.message);
                }
            },
            error: function() {
                alert('获取文件访问码失败');
            }
        });
    }
}

// 格式化文件大小
function formatSize(bytes) {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 Byte';
    const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
    return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i];
}

let loginIntervalId = null;

function startKeepLoginInterval() {
  // 如果已经存在，先清除
  if (loginIntervalId !== null) {
    clearInterval(loginIntervalId);
  }
  
  // 设置新的 interval 并保存 ID
  loginIntervalId = setInterval(keepLogin, 9 * 60 * 1000);
}


// Play a track
async function playTrackUrl(url) {
    // if (currentTrack) {
    //     const currentElement = document.querySelector('.playlist-item.active');
    //     if (currentElement) {
    //         currentElement.classList.remove('active');
    //     }
    // }

    // const element = Array.from(document.querySelectorAll('.playlist-item'))
    //     .find(el => el.textContent === track.name);
    // if (element) {
    //     element.classList.add('active');
    // }

    // currentTrack = track;
    
    try {
        await player.loadVideo({
            url: url,
            transport: "directfile",
            autoPlay: true,
        });
    } catch (error) {
        console.error('Error playing track:', error);
    }
}

var player = null;

// 页面加载完成后初始化
$(document).ready(function() {

    // 初始化 RxPlayer
    player = new RxPlayer({
        videoElement: document.getElementById("player"),
        audioTrackSwitchingMode: "direct",
    });
    // let currentTrack = null;



    const savedToken = localStorage.getItem('token');
    const savedUserInfo = localStorage.getItem('userInfo');
    if (savedToken) {
        token = savedToken;
        if (savedUserInfo) {
            currentUser = JSON.parse(savedUserInfo);
        }
        $('#loginForm').hide();
        $('#fileManager').show();
        
        // 显示用户信息
        updateUserInfo();

        // 检查 URL 中是否有 path 参数
        const urlParams = new URLSearchParams(window.location.search);
        const pathParam = urlParams.get('path');
        if (pathParam) {
            currentPath = decodeURIComponent(pathParam);
        }
        
        listFiles(currentPath);

        // 页面加载时获取部门和员工列表
        loadDeptList();
        loadEmpList();

        // 立即验证 token 并开始自动续期
         
        startKeepLoginInterval();
    } else {
        getCaptcha('login');
    }
    
    // 切换到注册表单
    $('#showRegister').click(function() {
        $('#loginForm').hide();
        $('#registerForm').show();
        getCaptcha('register');
    });

    // 切换到登录表单
    $('#showLogin').click(function() {
        $('#registerForm').hide();
        $('#loginForm').show();
        getCaptcha('login');
    });

    // 添加搜索框回车事件监听
    $('#searchInput').on('keypress', function(e) {
        if (e.which === 13) { // 13 是回车键的键码
            searchFiles();
        }
    });

    // 登录表单回车键处理
    $('#loginEmail').on('keypress', function(e) {
        if (e.which === 13) {
            $('#loginPassword').focus();
        }
    });

    $('#loginPassword').on('keypress', function(e) {
        if (e.which === 13) {
            $('#loginCaptcha').focus();
        }
    });

    $('#loginCaptcha').on('keypress', function(e) {
        if (e.which === 13) {
            login();
        }
    });

    // 点击验证码图片刷新验证码
    $('#captchaImage').on('click', function() {
        getCaptcha('login');
    });

    $('#regCaptchaImage').on('click', function() {
        getCaptcha('register');
    });

    // 检查 URL 中是否有搜索参数
    const urlParams = new URLSearchParams(window.location.search);
    const searchQuery = urlParams.get('search');
    if (searchQuery) {
        $('#searchInput').val(searchQuery);
        searchFiles(searchQuery);
    }

    // Initialize Media Player
    initializeMediaPlayer();
});

function initializeMediaPlayer() {
    const audioPlayer = document.getElementById('audioPlayer');
    const videoPlayer = document.getElementById('videoPlayer');
    const playPauseBtn = document.getElementById('playPause');
    const prevBtn = document.getElementById('prevTrack');
    const nextBtn = document.getElementById('nextTrack');
    const speedSelect = document.getElementById('playbackSpeed');
    const togglePlayModeBtn = document.getElementById('togglePlayMode');

    console.log("initializeMediaPlayer");

    // Play/Pause button
    playPauseBtn.addEventListener('click', () => {
        const activePlayer = getActivePlayer();
        if (activePlayer.paused) {
            activePlayer.play();
            updatePlayPauseIcon(true);
        } else {
            activePlayer.pause();
            updatePlayPauseIcon(false);
        }
    });

    // Previous track
    prevBtn.addEventListener('click', () => playPrevious());

    // Next track
    nextBtn.addEventListener('click', () => playNext());

    // Playback speed
    speedSelect.addEventListener('change', (e) => {
        const activePlayer = getActivePlayer();
        activePlayer.playbackRate = parseFloat(e.target.value);
    });

    // Toggle play mode
    togglePlayModeBtn.addEventListener('click', () => {
        isRandomPlay = !isRandomPlay;
        updatePlayModeIcon(isRandomPlay);
    });

    // Auto-play next track when current one ends
    audioPlayer.addEventListener('ended', () => playNext());
    videoPlayer.addEventListener('ended', () => playNext());

    // 添加播放状态变化的监听器
    audioPlayer.addEventListener('play', () => updatePlayPauseIcon(true));
    audioPlayer.addEventListener('pause', () => updatePlayPauseIcon(false));
    videoPlayer.addEventListener('play', () => updatePlayPauseIcon(true));
    videoPlayer.addEventListener('pause', () => updatePlayPauseIcon(false));
}

function getActivePlayer() {
    const audioPlayer = document.getElementById('audioPlayer');
    const videoPlayer = document.getElementById('videoPlayer');
    return videoPlayer.style.display === 'none' ? audioPlayer : videoPlayer;
}

function playNext() {
    if (playlist.length === 0) return;
    
    if (isRandomPlay) {
        currentTrackIndex = Math.floor(Math.random() * playlist.length);
    } else {
        currentTrackIndex = (currentTrackIndex + 1) % playlist.length;
    }
    playTrack(currentTrackIndex);
}

function playPrevious() {
    if (playlist.length === 0) return;
    
    if (isRandomPlay) {
        currentTrackIndex = Math.floor(Math.random() * playlist.length);
    } else {
        currentTrackIndex = (currentTrackIndex - 1 + playlist.length) % playlist.length;
    }
    playTrack(currentTrackIndex);
}



async function playTrack(index) {
    const track = playlist[index];
    const audioPlayer = document.getElementById('audioPlayer');
    const videoPlayer = document.getElementById('videoPlayer');
    const currentTrackSpan = document.getElementById('currentTrack');

    file_ext = track.path.split('.').pop();

    try {
        // First try to get the file access code
        const applyResponse = await fetch('/api/v1/files/apply', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                file: track.path
            })
        });
        
        if (!applyResponse.ok) {
            throw new Error('Failed to get file access code');
        }
        
        const applyData = await applyResponse.json();

        console.log(applyData)
        
        if (!applyData.code) {
            throw new Error('No access code returned: ' + applyData.message);
        }
        
        // Create the media URL with the access code
        const mediaUrl = `/api/v1/filesget/${encodeURIComponent(track.path)}?code=${applyData.code}&ext=.${file_ext}`;
        
        // Determine if it's a video file
        const isVideo = track.path.toLowerCase().endsWith('.mp4');
        
        // Show the appropriate player
        audioPlayer.style.display = isVideo ? 'none' : 'block';
        videoPlayer.style.display = isVideo ? 'block' : 'none';
        
        const activePlayer = isVideo ? videoPlayer : audioPlayer;
        
        // Set the source and play
        activePlayer.src = mediaUrl;
        
        // Add error handling for the player
        activePlayer.onerror = function(e) {
            console.error('Media error:', e);
            console.error('Error code:', activePlayer.error.code);
            console.error('Error message:', activePlayer.error.message);
            alert('播放错误: ' + activePlayer.error.message);
        };
        
        // Try to play the media
        const playPromise = activePlayer.play();
        
        if (playPromise !== undefined) {
            playPromise.catch(error => {
                console.error('Playback error:', error);
                alert('播放失败: ' + error.message);
            });
        }
        
        // Update UI
        updatePlayPauseIcon(true);
        currentTrackSpan.textContent = track.name;
        document.getElementById('mediaPlayer').style.display = 'block';
        
    } catch (error) {
        console.error('Error playing track:', error);
        alert('播放失败: ' + error.message);
    }
}


function playTrackGet(index) {
    const track = playlist[index];
    const audioPlayer = document.getElementById('audioPlayer');
    const videoPlayer = document.getElementById('videoPlayer');
    const currentTrackSpan = document.getElementById('currentTrack');

    file_ext = track.path.split('.').pop();

   

    // Get file access code and set up player
    $.ajax({
        url: '/api/v1/files/apply',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            file: track.path
        }),
        success: function(response) {
            if (response.code) {
                const mediaUrl = `/api/v1/filesget/${encodeURIComponent(track.path)}?code=${response.code}&ext=.${file_ext}`;
                const isVideo = track.path.toLowerCase().endsWith('.mp4');
                
                audioPlayer.style.display = isVideo ? 'none' : 'block';
                videoPlayer.style.display = isVideo ? 'block' : 'none';
                
                const activePlayer = isVideo ? videoPlayer : audioPlayer;
                activePlayer.src = mediaUrl;
                // activePlayer.play();

                playTrackUrl(mediaUrl);
                
                // 更新播放/暂停图标
                updatePlayPauseIcon(true);
                currentTrackSpan.textContent = track.name;
                document.getElementById('mediaPlayer').style.display = 'block';
            }
        }
    });
}

function updatePlayPauseIcon(isPlaying) {
    const playPauseBtn = document.getElementById('playPause');

    const playIcon = playPauseBtn.querySelector('.play-icon');
    const pauseIcon = playPauseBtn.querySelector('.pause-icon');
    playIcon.style.display = isPlaying ? 'none' : 'block';
    pauseIcon.style.display = isPlaying ? 'block' : 'none';
}

function updatePlayModeIcon(isRandom) {
    const togglePlayModeBtn = document.getElementById('togglePlayMode');

    const sequenceIcon = togglePlayModeBtn.querySelector('.sequence-icon');
    const randomIcon = togglePlayModeBtn.querySelector('.random-icon');
    sequenceIcon.style.display = isRandom ? 'none' : 'block';
    randomIcon.style.display = isRandom ? 'block' : 'none';
}

// 文件点赞/不喜欢功能
function likeFile(path, degree) {
    $.ajax({
        url: '/api/v1/files/like',
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        contentType: 'application/json',
        data: JSON.stringify({
            file: path,
            degree: degree
        }),
        success: function(response) {
            // 检查response是否存在
            if (!response) {
                showToast('操作失败: 服务器无响应');
                return;
            }
            
            if (response.code === 200) {
                // 显示操作成功提示
                let message = '';
                switch(degree) {
                    case 5:
                        message = '已标记为 Obsessed (超级着迷的)';
                        break;
                    case 4:
                        message = '已标记为 Awesome (超棒的)';
                        break;
                    case 3:
                        message = '已标记为 Enjoyable (令人享受的)';
                        break;
                    case 2:
                        message = '已标记为 Catchy (朗朗上口，有吸引力)';
                        break;
                    case 1:
                        message = '已标记为 Nice (还不错)';
                        break;
                    default:
                        message = '已添加到喜欢';
                }
                showToast(message);
            } else if (response.code === 401) {
                alert('登录已过期，请重新登录');
                logout();
            } else {
                // 确保response.message存在
                const errorMsg = response.message || '未知错误';
                showToast('操作失败: ' + errorMsg);
            }
        },
        error: function(xhr, status, error) {
            console.error('Like file error:', status, error);
            showToast('操作失败，请稍后重试');
        }
    });
}

// 显示提示消息
function showToast(message) {
    // 检查是否已存在 toast 元素
    let toast = $('#toast');
    if (toast.length === 0) {
        // 如果不存在，创建一个新的 toast 元素
        toast = $('<div id="toast" class="toast"></div>');
        $('body').append(toast);
    }
    
    // 设置消息内容并显示
    toast.text(message).fadeIn();
    
    // 3秒后自动隐藏
    setTimeout(function() {
        toast.fadeOut();
    }, 3000);
}
