/**
 * Ticket Kanban Board - Main Frontend Logic (Obsidian/Linear Style)
 */

// i18n Translations
const translations = {
    zh: {
        search_placeholder: "搜索 ID 或标题...",
        all_categories: "所有分类",
        all_priorities: "所有优先级",
        all_owners: "所有负责人",
        btn_refresh: "刷新",
        btn_new_ticket: "新建工单",
        status_open: "待处理",
        status_fixing: "处理中",
        status_resolved: "已解决",
        status_passed: "已关闭",
        status_rejected: "已拒绝",
        tab_write: "编辑描述",
        tab_preview: "预览效果",
        placeholder_title: "工单标题",
        placeholder_body: "在这里使用 Markdown 描述你的任务或缺陷...",
        sidebar_title: "工单属性",
        label_status: "状态",
        opt_status_open: "待处理 (Open)",
        opt_status_fixing: "处理中 (Fixing)",
        opt_status_resolved: "已解决 (Resolved)",
        opt_status_passed: "已验证/关闭 (Passed)",
        opt_status_rejected: "已拒绝 (Rejected)",
        label_priority: "优先级",
        label_owner: "负责人",
        placeholder_owner: "输入负责人姓名或账号",
        label_conclusion: "解决方案 (结论)",
        placeholder_conclusion: "对于已解决/拒绝的工单，在此填写最终解决结论...",
        label_created_at: "创建时间",
        label_updated_at: "更新时间",
        label_resolved_at: "解决时间",
        btn_delete: "删除工单",
        btn_cancel: "取消",
        btn_save: "保存修改",
        title_create_ticket: "新建工单",
        label_ticket_title: "工单标题",
        placeholder_create_title: "输入简洁明了的工单标题",
        label_detailed_desc: "详细描述",
        placeholder_create_body: "在这里描述工单的具体细节，支持 Markdown...",
        label_category: "工单分类/目录",
        label_ticket_id: "工单 ID",
        placeholder_ticket_id: "如: BUG-123",
        placeholder_owner_simple: "工单负责人",
        label_initial_status: "初始状态",
        btn_create_now: "立即创建",
        msg_saving: "正在保存...",
        msg_saved: "保存成功！",
        msg_deleting: "正在删除...",
        msg_deleted: "删除成功！",
        msg_creating: "正在创建...",
        msg_created: "创建成功！",
        msg_load_failed: "加载失败",
        msg_save_failed: "保存失败",
        msg_delete_failed: "删除失败",
        msg_create_failed: "创建失败",
        msg_confirm_delete: "确定要彻底删除工单 {id} 对应的 Markdown 文件吗？\n此操作不可逆！",
        msg_fill_required: "请填写所有必填字段！",
        msg_invalid_id: "ID 只能包含字母、数字和连字符",
        msg_id_exists: "ID 已存在",
        msg_error: "发生错误",
        msg_success: "成功",
        msg_no_tickets: "暂无工单，请先新建！",
        btn_refresh_title: "从文件系统重新加载"
    },
    en: {
        search_placeholder: "Search ID or title...",
        all_categories: "All Categories",
        all_priorities: "All Priorities",
        all_owners: "All Owners",
        btn_refresh: "Refresh",
        btn_new_ticket: "New Ticket",
        status_open: "Open",
        status_fixing: "Fixing",
        status_resolved: "Resolved",
        status_passed: "Closed",
        status_rejected: "Rejected",
        tab_write: "Write",
        tab_preview: "Preview",
        placeholder_title: "Ticket Title",
        placeholder_body: "Use Markdown to describe the task or bug here...",
        sidebar_title: "Ticket Attributes",
        label_status: "Status",
        opt_status_open: "Open",
        opt_status_fixing: "Fixing",
        opt_status_resolved: "Resolved",
        opt_status_passed: "Closed (Passed)",
        opt_status_rejected: "Rejected",
        label_priority: "Priority",
        label_owner: "Owner",
        placeholder_owner: "Enter owner name or username",
        label_conclusion: "Resolution (Conclusion)",
        placeholder_conclusion: "For resolved/rejected tickets, write the conclusion here...",
        label_created_at: "Created At",
        label_updated_at: "Updated At",
        label_resolved_at: "Resolved At",
        btn_delete: "Delete Ticket",
        btn_cancel: "Cancel",
        btn_save: "Save Changes",
        title_create_ticket: "Create Ticket",
        label_ticket_title: "Title",
        placeholder_create_title: "Enter a brief title",
        label_detailed_desc: "Description",
        placeholder_create_body: "Describe the ticket details, Markdown supported...",
        label_category: "Category / Directory",
        label_ticket_id: "Ticket ID",
        placeholder_ticket_id: "e.g. BUG-123",
        placeholder_owner_simple: "Ticket owner",
        label_initial_status: "Initial Status",
        btn_create_now: "Create Now",
        msg_saving: "Saving...",
        msg_saved: "Saved successfully!",
        msg_deleting: "Deleting...",
        msg_deleted: "Deleted successfully!",
        msg_creating: "Creating...",
        msg_created: "Created successfully!",
        msg_load_failed: "Failed to load",
        msg_save_failed: "Failed to save",
        msg_delete_failed: "Failed to delete",
        msg_create_failed: "Failed to create",
        msg_confirm_delete: "Are you sure you want to delete ticket {id}? This action cannot be undone!",
        msg_fill_required: "Please fill in all required fields!",
        msg_invalid_id: "ID can only contain letters, numbers, and hyphens",
        msg_id_exists: "ID already exists",
        msg_error: "Error occurred",
        msg_success: "Success",
        msg_no_tickets: "No tickets found. Create one to start!",
        btn_refresh_title: "Reload from file system"
    }
};

let currentLang = 'en';

function t(key) {
    return translations[currentLang][key] || key;
}

function initLanguage() {
    // Detect language preference (localStorage -> browser lang -> default en)
    const savedLang = localStorage.getItem('ticket-lang');
    if (savedLang && translations[savedLang]) {
        currentLang = savedLang;
    } else {
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang && browserLang.startsWith('zh')) {
            currentLang = 'zh';
        } else {
            currentLang = 'en';
        }
    }
    applyLanguage(currentLang);
}

function applyLanguage(lang) {
    currentLang = lang;
    localStorage.setItem('ticket-lang', lang);

    // Update elements with data-i18n attributes
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (translations[lang][key]) {
            el.textContent = translations[lang][key];
        }
    });

    // Update placeholders with data-i18n-placeholder
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (translations[lang][key]) {
            el.setAttribute('placeholder', translations[lang][key]);
        }
    });

    // Update titles with data-i18n-title
    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        const key = el.getAttribute('data-i18n-title');
        if (translations[lang][key]) {
            el.setAttribute('title', translations[lang][key]);
        }
    });

    // Toggle language button label
    const btnLang = document.getElementById('btn-lang');
    if (btnLang) {
        btnLang.textContent = lang === 'en' ? '🌐 ZH' : '🌐 EN';
    }
}

// Global State
const state = {
    tickets: [],
    config: {
        sub_dirs: [],
        extra_fields: []
    },
    activeTicket: null, // Holds detailed info (meta + body)
    filters: {
        search: '',
        type: '',
        priority: '',
        owner: ''
    }
};

// DOM Elements
const elements = {
    // Kanban Columns
    columns: document.querySelectorAll('.kanban-column'),
    cardsContainers: {
        open: document.getElementById('cards-open'),
        fixing: document.getElementById('cards-fixing'),
        resolved: document.getElementById('cards-resolved'),
        passed: document.getElementById('cards-passed'),
        rejected: document.getElementById('cards-rejected')
    },
    counts: {
        open: document.getElementById('count-open'),
        fixing: document.getElementById('count-fixing'),
        resolved: document.getElementById('count-resolved'),
        passed: document.getElementById('count-passed'),
        rejected: document.getElementById('count-rejected')
    },
    
    // Filters & Controls
    searchInput: document.getElementById('search-input'),
    filterType: document.getElementById('filter-type'),
    filterPriority: document.getElementById('filter-priority'),
    filterOwner: document.getElementById('filter-owner'),
    btnNewTicket: document.getElementById('btn-new-ticket'),
    btnRefresh: document.getElementById('btn-refresh'),
    btnLang: document.getElementById('btn-lang'),
    
    // Detail Modal
    modalDetail: document.getElementById('modal-detail'),
    detailBadgeId: document.getElementById('detail-badge-id'),
    detailBadgeType: document.getElementById('detail-badge-type'),
    detailTitle: document.getElementById('detail-title'),
    detailBody: document.getElementById('detail-body'),
    detailPreview: document.getElementById('detail-preview'),
    detailStatus: document.getElementById('detail-status'),
    detailPriority: document.getElementById('detail-priority'),
    detailOwner: document.getElementById('detail-owner'),
    detailConclusion: document.getElementById('detail-conclusion'),
    detailCreatedAt: document.getElementById('detail-created-at'),
    detailUpdatedAt: document.getElementById('detail-updated-at'),
    detailResolvedAt: document.getElementById('detail-resolved-at'),
    detailResolvedContainer: document.getElementById('detail-resolved-container'),
    detailBtnSave: document.getElementById('detail-btn-save'),
    detailBtnCancel: document.getElementById('detail-btn-cancel'),
    detailBtnDelete: document.getElementById('detail-btn-delete'),
    detailBtnClose: document.getElementById('detail-btn-close'),
    tabWrite: document.getElementById('tab-write'),
    tabPreview: document.getElementById('tab-preview'),
    dynamicFieldsContainer: document.getElementById('dynamic-fields-container'),
    
    // Create Modal
    modalCreate: document.getElementById('modal-create'),
    formCreate: document.getElementById('form-create'),
    createTitle: document.getElementById('create-title'),
    createBody: document.getElementById('create-body'),
    createDir: document.getElementById('create-dir'),
    createId: document.getElementById('create-id'),
    createPriority: document.getElementById('create-priority'),
    createOwner: document.getElementById('create-owner'),
    createStatus: document.getElementById('create-status'),
    btnGenerateId: document.getElementById('btn-generate-id'),
    createBtnCancel: document.getElementById('create-btn-cancel'),
    createBtnClose: document.getElementById('create-btn-close'),
    createDynamicFieldsContainer: document.getElementById('create-dynamic-fields-container'),
    
    // Toasts
    toastContainer: document.getElementById('toast-container')
};

// Initialize App
document.addEventListener('DOMContentLoaded', () => {
    initLanguage();
    fetchConfig().then(() => {
        fetchTickets();
    });
    setupEventListeners();
    setupDragAndDrop();
});

// ==========================================================================
// API Interaction & Fetching
// ==========================================================================

async function fetchConfig() {
    try {
        const response = await fetch('/api/config');
        if (!response.ok) throw new Error(t('msg_load_failed'));
        state.config = await response.json();
        
        // Populate Filters and Creation Directory selectors
        populateConfigSelectors();
    } catch (err) {
        showToast(t('msg_load_failed') + ': ' + err.message, 'error');
    }
}

async function fetchTickets() {
    try {
        const response = await fetch('/api/tickets');
        if (!response.ok) throw new Error(t('msg_load_failed'));
        state.tickets = await response.json();
        
        // Populate Owner Filter list with unique owners
        populateOwnerFilter();
        
        // Render Columns
        renderKanban();
    } catch (err) {
        showToast(t('msg_load_failed') + ': ' + err.message, 'error');
    }
}

function populateConfigSelectors() {
    // Populate directory select in "Create Modal" and type filter
    elements.createDir.innerHTML = '';
    elements.filterType.innerHTML = `<option value="">${t('all_categories')}</option>`;
    
    if (state.config.sub_dirs && state.config.sub_dirs.length > 0) {
        state.config.sub_dirs.forEach(sub => {
            // Option for Creation form
            const optCreate = document.createElement('option');
            optCreate.value = sub;
            optCreate.textContent = sub;
            elements.createDir.appendChild(optCreate);
            
            // Option for filter
            const optFilter = document.createElement('option');
            optFilter.value = sub;
            optFilter.textContent = sub;
            elements.filterType.appendChild(optFilter);
        });
    } else {
        // Fallbacks
        const optDefault = document.createElement('option');
        optDefault.value = 'tickets';
        optDefault.textContent = 'tickets (' + (currentLang === 'zh' ? '默认' : 'default') + ')';
        elements.createDir.appendChild(optDefault);
    }
}

function populateOwnerFilter() {
    const owners = new Set();
    state.tickets.forEach(t => {
        if (t.meta && t.meta.owner) {
            owners.add(t.meta.owner.trim());
        }
    });
    
    // Save current selected value
    const currentSelected = elements.filterOwner.value;
    elements.filterOwner.innerHTML = `<option value="">${t('all_owners')}</option>`;
    
    Array.from(owners).sort().forEach(owner => {
        const opt = document.createElement('option');
        opt.value = owner;
        opt.textContent = owner;
        elements.filterOwner.appendChild(opt);
    });
    
    // Restore selection
    elements.filterOwner.value = currentSelected;
}

// ==========================================================================
// Kanban Rendering & Filters
// ==========================================================================

function renderKanban() {
    // Clear lists
    Object.values(elements.cardsContainers).forEach(container => {
        container.innerHTML = '';
    });
    
    // Columns count tracking
    const columnCounts = { open: 0, fixing: 0, resolved: 0, passed: 0, rejected: 0 };
    
    // Filter tickets
    const filteredTickets = state.tickets.filter(ticket => {
        const meta = ticket.meta || {};
        const title = (meta.title || '').toLowerCase();
        const id = (meta.id || '').toLowerCase();
        const searchWord = state.filters.search.toLowerCase();
        
        // Search filter (matches ID, title)
        if (searchWord && !title.includes(searchWord) && !id.includes(searchWord)) {
            return false;
        }
        
        // Type filter (directory folder name)
        if (state.filters.type && meta.type !== state.filters.type) {
            return false;
        }
        
        // Priority filter
        if (state.filters.priority && meta.priority !== state.filters.priority) {
            return false;
        }
        
        // Owner filter
        if (state.filters.owner && meta.owner !== state.filters.owner) {
            return false;
        }
        
        return true;
    });

    // Populate columns
    filteredTickets.forEach(ticket => {
        const status = (ticket.meta.status || 'open').toLowerCase();
        const container = elements.cardsContainers[status];
        if (container) {
            container.appendChild(createTicketCard(ticket));
            columnCounts[status]++;
        }
    });
    
    // Update counters
    Object.keys(elements.counts).forEach(status => {
        elements.counts[status].textContent = columnCounts[status];
    });
}

function createTicketCard(ticket) {
    const meta = ticket.meta || {};
    const card = document.createElement('div');
    card.className = 'ticket-card';
    card.setAttribute('draggable', 'true');
    card.setAttribute('data-path', ticket.file_path);
    card.setAttribute('data-priority', meta.priority || 'minor');
    
    // Priority labels
    const priorityLabel = (meta.priority || 'minor').toUpperCase();
    
    // Short title to avoid text overflows
    const title = meta.title || 'Untitled Ticket';
    
    // Get Avatar initials
    const ownerName = meta.owner || (currentLang === 'zh' ? '未指派' : 'Unassigned');
    const avatarInitials = ownerName.trim().slice(0, 2).toUpperCase();
    
    card.innerHTML = `
        <div class="card-top">
            <span class="card-id">${meta.id}</span>
            <span class="card-type-tag">${meta.type || 'task'}</span>
        </div>
        <div class="card-title">${escapeHTML(title)}</div>
        <div class="card-footer">
            <div class="card-footer-left">
                <span class="priority-badge priority-${meta.priority || 'minor'}">${priorityLabel}</span>
                <span class="card-owner" title="${currentLang === 'zh' ? '负责人' : 'Owner'}: ${escapeHTML(ownerName)}">
                    <span class="owner-avatar">${escapeHTML(avatarInitials)}</span>
                </span>
            </div>
            <span class="card-time">${formatDate(meta.updated_at || meta.created_at)}</span>
        </div>
    `;
    
    // Card Click open detail
    card.addEventListener('click', () => {
        openDetailModal(ticket.file_path);
    });
    
    // Setup drag elements
    card.addEventListener('dragstart', (e) => {
        card.classList.add('dragging');
        e.dataTransfer.setData('text/plain', ticket.file_path);
        e.dataTransfer.effectAllowed = 'move';
    });
    
    card.addEventListener('dragend', () => {
        card.classList.remove('dragging');
    });
    
    return card;
}

// ==========================================================================
// Drag and Drop Logic
// ==========================================================================

function setupDragAndDrop() {
    elements.columns.forEach(column => {
        column.addEventListener('dragover', (e) => {
            e.preventDefault();
            column.classList.add('drag-over');
        });
        
        column.addEventListener('dragleave', () => {
            column.classList.remove('drag-over');
        });
        
        column.addEventListener('drop', async (e) => {
            e.preventDefault();
            column.classList.remove('drag-over');
            
            const filePath = e.dataTransfer.getData('text/plain');
            const newStatus = column.getAttribute('data-status');
            
            // Find target card in state locally to prevent rendering delay
            const ticketIndex = state.tickets.findIndex(t => t.file_path === filePath);
            if (ticketIndex === -1) return;
            
            const oldStatus = state.tickets[ticketIndex].meta.status;
            if (oldStatus === newStatus) return; // No change
            
            // Optimistic Update locally
            state.tickets[ticketIndex].meta.status = newStatus;
            state.tickets[ticketIndex].meta.updated_at = formatNowDate();
            renderKanban();
            
            // Send API call to save on server
            try {
                const response = await fetch('/api/tickets/move', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        file_path: filePath,
                        status: newStatus
                    })
                });
                
                if (!response.ok) {
                    throw new Error(currentLang === 'zh' ? '更新状态失败' : 'Failed to update status');
                }
                
                // Fetch latest data to reconcile metadata (resolved_at, etc.)
                const updatedTicket = await response.json();
                state.tickets[ticketIndex].meta = updatedTicket.meta;
                renderKanban();
                
                const msg = currentLang === 'zh' ? 
                    `已移动工单 ${updatedTicket.meta.id} 至 ${newStatus.toUpperCase()}` : 
                    `Moved ticket ${updatedTicket.meta.id} to ${newStatus.toUpperCase()}`;
                showToast(msg, 'success');
            } catch (err) {
                // Rollback on failure
                state.tickets[ticketIndex].meta.status = oldStatus;
                renderKanban();
                showToast((currentLang === 'zh' ? '更改状态失败: ' : 'Failed to change status: ') + err.message, 'error');
            }
        });
    });
}

// ==========================================================================
// Detail Modal View & Edit
// ==========================================================================

async function openDetailModal(filePath) {
    try {
        const response = await fetch(`/api/tickets/detail?path=${encodeURIComponent(filePath)}`);
        if (!response.ok) throw new Error(t('msg_load_failed'));
        
        state.activeTicket = await response.json();
        const ticket = state.activeTicket;
        const meta = ticket.meta || {};
        
        // Reset tabs to Editor default
        switchDetailTab('write');
        
        // Setup fields
        elements.detailBadgeId.textContent = meta.id;
        elements.detailBadgeType.textContent = (meta.type || 'task').toUpperCase();
        elements.detailTitle.value = meta.title || '';
        elements.detailBody.value = ticket.body || '';
        elements.detailStatus.value = meta.status || 'open';
        elements.detailPriority.value = meta.priority || 'minor';
        elements.detailOwner.value = meta.owner || '';
        elements.detailConclusion.value = meta.conclusion || '';
        elements.detailCreatedAt.textContent = meta.created_at || '-';
        elements.detailUpdatedAt.textContent = meta.updated_at || '-';
        
        if (meta.resolved_at) {
            elements.detailResolvedAt.textContent = meta.resolved_at;
            elements.detailResolvedContainer.classList.remove('hidden');
        } else {
            elements.detailResolvedContainer.classList.add('hidden');
        }
        
        // Generate extra metadata fields dynamically
        renderDynamicFields(meta);
        
        // Show Modal
        elements.modalDetail.classList.remove('hidden');
    } catch (err) {
        showToast((currentLang === 'zh' ? '无法打开工单详情: ' : 'Cannot open ticket details: ') + err.message, 'error');
    }
}

function renderDynamicFields(meta) {
    elements.dynamicFieldsContainer.innerHTML = '';
    const extraFields = state.config.extra_fields || [];
    
    if (extraFields.length === 0) return;
    
    // Title
    const title = document.createElement('h4');
    title.className = 'sidebar-title';
    title.style.marginTop = '16px';
    title.textContent = currentLang === 'zh' ? '自定义字段' : 'Custom Fields';
    elements.dynamicFieldsContainer.appendChild(title);
    
    extraFields.forEach(field => {
        const container = document.createElement('div');
        container.className = 'meta-field';
        
        // Find existing value
        let val = '';
        if (meta.extra_fields && meta.extra_fields[field.name] !== undefined) {
            val = meta.extra_fields[field.name];
        } else if (meta[field.name] !== undefined && typeof meta[field.name] !== 'object') {
            val = meta[field.name];
        }
        
        container.innerHTML = `
            <label for="dynamic-field-${field.name}">${field.name}${field.required ? ' *' : ''}</label>
            <input type="text" id="dynamic-field-${field.name}" data-field-name="${field.name}" value="${escapeHTML(String(val))}" placeholder="${currentLang === 'zh' ? '填写' : 'Enter'} ${field.name}" ${field.required ? 'required' : ''}>
        `;
        elements.dynamicFieldsContainer.appendChild(container);
    });
}

function switchDetailTab(tab) {
    if (tab === 'write') {
        elements.tabWrite.classList.add('active');
        elements.tabPreview.classList.remove('active');
        elements.detailBody.classList.remove('hidden');
        elements.detailPreview.classList.add('hidden');
    } else {
        elements.tabWrite.classList.remove('active');
        elements.tabPreview.classList.add('active');
        elements.detailBody.classList.add('hidden');
        elements.detailPreview.classList.remove('hidden');
        
        // Render Markdown
        const mdText = elements.detailBody.value || `*${currentLang === 'zh' ? '无描述' : 'No description'}*`;
        if (window.marked && typeof window.marked.parse === 'function') {
            elements.detailPreview.innerHTML = window.marked.parse(mdText);
        } else {
            elements.detailPreview.innerHTML = `<pre style="white-space: pre-wrap; font-family: sans-serif;">${escapeHTML(mdText)}</pre>`;
        }
    }
}

async function saveTicketDetail() {
    if (!state.activeTicket) return;
    
    const filePath = state.activeTicket.file_path;
    const title = elements.detailTitle.value.trim();
    if (!title) {
        showToast(t('msg_fill_required'), 'error');
        return;
    }
    
    // Standard payload
    const updates = {
        title: title,
        status: elements.detailStatus.value,
        priority: elements.detailPriority.value,
        owner: elements.detailOwner.value.trim(),
        conclusion: elements.detailConclusion.value.trim()
    };
    
    // Collect dynamic extra fields
    const dynamicInputs = elements.dynamicFieldsContainer.querySelectorAll('input[data-field-name]');
    let hasValidationErr = false;
    
    dynamicInputs.forEach(input => {
        const fieldName = input.getAttribute('data-field-name');
        const val = input.value.trim();
        const required = input.hasAttribute('required');
        
        if (required && !val) {
            const msg = currentLang === 'zh' ? `必填字段 '${fieldName}' 不能为空` : `Required field '${fieldName}' cannot be empty`;
            showToast(msg, 'error');
            hasValidationErr = true;
            input.focus();
        }
        
        updates[fieldName] = val;
    });
    
    if (hasValidationErr) return;
    
    const bodyText = elements.detailBody.value;
    
    try {
        elements.detailBtnSave.disabled = true;
        elements.detailBtnSave.textContent = t('msg_saving');
        
        const response = await fetch('/api/tickets/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                file_path: filePath,
                updates: updates,
                body: bodyText
            })
        });
        
        if (!response.ok) {
            const errData = await response.json().catch(() => ({}));
            throw new Error(errData.error || t('msg_save_failed'));
        }
        
        showToast(t('msg_saved'), 'success');
        elements.modalDetail.classList.add('hidden');
        state.activeTicket = null;
        
        // Refresh Lists
        fetchTickets();
    } catch (err) {
        showToast((currentLang === 'zh' ? '更新失败: ' : 'Failed to update: ') + err.message, 'error');
    } finally {
        elements.detailBtnSave.disabled = false;
        elements.detailBtnSave.textContent = t('btn_save');
    }
}

async function deleteActiveTicket() {
    if (!state.activeTicket) return;
    const ticketId = state.activeTicket.meta.id;
    const filePath = state.activeTicket.file_path;
    
    if (!confirm(t('msg_confirm_delete').replace('{id}', ticketId))) {
        return;
    }
    
    try {
        const response = await fetch(`/api/tickets/delete`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ file_path: filePath })
        });
        
        if (!response.ok) throw new Error(t('msg_delete_failed'));
        
        const msg = currentLang === 'zh' ? `工单 ${ticketId} 已成功删除` : `Ticket ${ticketId} deleted successfully`;
        showToast(msg, 'success');
        elements.modalDetail.classList.add('hidden');
        state.activeTicket = null;
        
        fetchTickets();
    } catch (err) {
        showToast((currentLang === 'zh' ? '删除工单失败: ' : 'Failed to delete ticket: ') + err.message, 'error');
    }
}

// ==========================================================================
// Create New Ticket Form
// ==========================================================================

function openCreateModal() {
    // Reset Form
    elements.formCreate.reset();
    elements.createBody.value = '';
    
    // Populate dynamic extra fields
    renderCreateDynamicFields();
    
    // Auto recommend ID
    recommendNextId();
    
    elements.modalCreate.classList.remove('hidden');
    elements.createTitle.focus();
}

function renderCreateDynamicFields() {
    elements.createDynamicFieldsContainer.innerHTML = '';
    const extraFields = state.config.extra_fields || [];
    
    if (extraFields.length === 0) return;
    
    extraFields.forEach(field => {
        const row = document.createElement('div');
        row.className = 'form-row';
        row.innerHTML = `
            <label for="create-dynamic-${field.name}" class="${field.required ? 'required-label' : ''}">${field.name}</label>
            <input type="text" id="create-dynamic-${field.name}" data-field-name="${field.name}" placeholder="${currentLang === 'zh' ? '输入' : 'Enter'} ${field.name}" ${field.required ? 'required' : ''}>
        `;
        elements.createDynamicFieldsContainer.appendChild(row);
    });
}

function recommendNextId() {
    const selectedDir = elements.createDir.value || 'bugs';
    let prefix = 'TASK';
    if (selectedDir.toLowerCase().includes('bug')) {
        prefix = 'BUG';
    } else if (selectedDir.toLowerCase().includes('task')) {
        prefix = 'TASK';
    } else if (selectedDir.toLowerCase().includes('feature')) {
        prefix = 'FEAT';
    } else {
        prefix = selectedDir.slice(0, 4).toUpperCase();
    }
    
    // Scan existing IDs starting with this prefix
    let maxNum = 0;
    state.tickets.forEach(t => {
        const id = t.meta.id || '';
        if (id.toUpperCase().startsWith(prefix + '-')) {
            const numPart = id.slice(prefix.length + 1);
            const num = parseInt(numPart, 10);
            if (!isNaN(num) && num > maxNum) {
                maxNum = num;
            }
        }
    });
    
    const nextNum = maxNum + 1;
    const paddedNum = String(nextNum).padStart(2, '0');
    elements.createId.value = `${prefix}-${paddedNum}`;
}

async function handleCreateSubmit(e) {
    e.preventDefault();
    
    const title = elements.createTitle.value.trim();
    const id = elements.createId.value.trim();
    const dir = elements.createDir.value;
    
    if (!title || !id || !dir) {
        showToast(t('msg_fill_required'), 'error');
        return;
    }
    
    // Basic fields
    const payload = {
        dir: dir,
        id: id,
        title: title,
        status: elements.createStatus.value,
        priority: elements.createPriority.value,
        owner: elements.createOwner.value.trim(),
        body: elements.createBody.value,
        extra_fields: {}
    };
    
    // Collect dynamic extra fields
    const dynamicInputs = elements.createDynamicFieldsContainer.querySelectorAll('input[data-field-name]');
    let hasValidationErr = false;
    
    dynamicInputs.forEach(input => {
        const fieldName = input.getAttribute('data-field-name');
        const val = input.value.trim();
        const required = input.hasAttribute('required');
        
        if (required && !val) {
            const msg = currentLang === 'zh' ? `字段 '${fieldName}' 必填！` : `Field '${fieldName}' is required!`;
            showToast(msg, 'error');
            hasValidationErr = true;
            input.focus();
        }
        
        payload.extra_fields[fieldName] = val;
    });
    
    if (hasValidationErr) return;
    
    try {
        const submitBtn = document.getElementById('create-btn-submit');
        submitBtn.disabled = true;
        submitBtn.textContent = t('msg_creating');
        
        const response = await fetch('/api/tickets/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            const errData = await response.json().catch(() => ({}));
            throw new Error(errData.error || t('msg_create_failed'));
        }
        
        showToast(t('msg_created').replace('{id}', id), 'success');
        elements.modalCreate.classList.add('hidden');
        
        // Reload list
        fetchTickets();
    } catch (err) {
        showToast((currentLang === 'zh' ? '创建失败: ' : 'Failed to create: ') + err.message, 'error');
    } finally {
        const submitBtn = document.getElementById('create-btn-submit');
        submitBtn.disabled = false;
        submitBtn.textContent = t('btn_create_now');
    }
}

// ==========================================================================
// UI Helpers & Setup Event Listeners
// ==========================================================================

function setupEventListeners() {
    // Language toggle button click
    if (elements.btnLang) {
        elements.btnLang.addEventListener('click', () => {
            const targetLang = currentLang === 'en' ? 'zh' : 'en';
            applyLanguage(targetLang);
            
            // Re-populate selectors since they contain dynamically translated options
            populateConfigSelectors();
            populateOwnerFilter();
            
            // Render kanban because tickets render dynamic owner title tooltips, etc.
            renderKanban();
        });
    }

    // Search filter event
    elements.searchInput.addEventListener('input', (e) => {
        state.filters.search = e.target.value.trim();
        renderKanban();
    });
    
    // Select filter events
    elements.filterType.addEventListener('change', (e) => {
        state.filters.type = e.target.value;
        renderKanban();
    });
    elements.filterPriority.addEventListener('change', (e) => {
        state.filters.priority = e.target.value;
        renderKanban();
    });
    elements.filterOwner.addEventListener('change', (e) => {
        state.filters.owner = e.target.value;
        renderKanban();
    });
    
    // Modal buttons actions
    elements.btnNewTicket.addEventListener('click', openCreateModal);
    elements.btnRefresh.addEventListener('click', () => {
        fetchTickets();
        const msg = currentLang === 'zh' ? '工单列表已刷新' : 'Ticket list refreshed';
        showToast(msg, 'success');
    });
    
    // Detail Modal actions
    elements.detailBtnClose.addEventListener('click', () => { elements.modalDetail.classList.add('hidden'); });
    elements.detailBtnCancel.addEventListener('click', () => { elements.modalDetail.classList.add('hidden'); });
    elements.detailBtnSave.addEventListener('click', saveTicketDetail);
    elements.detailBtnDelete.addEventListener('click', deleteActiveTicket);
    
    // Tab switcher actions
    elements.tabWrite.addEventListener('click', () => switchDetailTab('write'));
    elements.tabPreview.addEventListener('click', () => switchDetailTab('preview'));
    
    // Create Modal actions
    elements.createBtnClose.addEventListener('click', () => { elements.modalCreate.classList.add('hidden'); });
    elements.createBtnCancel.addEventListener('click', () => { elements.modalCreate.classList.add('hidden'); });
    elements.formCreate.addEventListener('submit', handleCreateSubmit);
    
    // Recommend ID when dir changes
    elements.createDir.addEventListener('change', recommendNextId);
    elements.btnGenerateId.addEventListener('click', recommendNextId);
    
    // Close modal on click overlay
    window.addEventListener('click', (e) => {
        if (e.target === elements.modalDetail) elements.modalDetail.classList.add('hidden');
        if (e.target === elements.modalCreate) elements.modalCreate.classList.add('hidden');
    });
}

// Toast Notifications manager
function showToast(msg, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    
    // Simple inline icon representation
    let iconSvg = '';
    if (type === 'success') {
        iconSvg = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--status-resolved-text)" stroke-width="2" stroke-linecap="round"><polyline points="20 6 9 17 4 12"></polyline></svg>`;
    } else if (type === 'error') {
        iconSvg = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--status-rejected-text)" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`;
    } else {
        iconSvg = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--primary-color)" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>`;
    }
    
    toast.innerHTML = `
        ${iconSvg}
        <span>${escapeHTML(msg)}</span>
    `;
    
    elements.toastContainer.appendChild(toast);
    
    // Fade out and remove
    setTimeout(() => {
        toast.classList.add('fade-out');
        setTimeout(() => {
            toast.remove();
        }, 300);
    }, 3500);
}

// Escapes raw strings for HTML representation
function escapeHTML(str) {
    if (!str) return '';
    return str.replace(/[&<>'"]/g, 
        tag => ({
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            "'": '&#39;',
            '"': '&quot;'
        }[tag] || tag)
    );
}

// Date formatter
function formatDate(dateStr) {
    if (!dateStr) return '';
    // Format "2006-01-02 15:04" -> "01/02 15:04"
    if (dateStr.length >= 16) {
        return dateStr.slice(5);
    }
    return dateStr;
}

function formatNowDate() {
    const now = new Date();
    const pad = (n) => String(n).padStart(2, '0');
    return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}`;
}
