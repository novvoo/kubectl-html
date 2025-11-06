package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// K8s 资源结构
type K8sResource struct {
	APIVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata" json:"metadata"`
	Spec       interface{}            `yaml:"spec,omitempty" json:"spec,omitempty"`
	Status     interface{}            `yaml:"status,omitempty" json:"status,omitempty"`
	Data       interface{}            `yaml:"data,omitempty" json:"data,omitempty"`
	StringData interface{}            `yaml:"stringData,omitempty" json:"stringData,omitempty"`
	Rules      interface{}            `yaml:"rules,omitempty" json:"rules,omitempty"`
	Subjects   interface{}            `yaml:"subjects,omitempty" json:"subjects,omitempty"`
	RoleRef    interface{}            `yaml:"roleRef,omitempty" json:"roleRef,omitempty"`
}

type K8sList struct {
	APIVersion string        `yaml:"apiVersion" json:"apiVersion"`
	Kind       string        `yaml:"kind" json:"kind"`
	Items      []K8sResource `yaml:"items" json:"items"`
}

type ResourceInfo struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Kind       string                 `json:"kind"`
	APIVersion string                 `json:"apiVersion"`
	Age        string                 `json:"age"`
	Status     string                 `json:"status"`
	YAML       string                 `json:"yaml"`
	Parsed     map[string]interface{} `json:"parsed"`
}

// HTML 模板（内嵌）
const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>kubectl html - {{ .Command }}</title>
  <style>
    * { box-sizing: border-box; }
    body { 
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
      margin: 0; padding: 20px; 
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
    }
    .container { 
      max-width: 1400px; margin: 0 auto; 
      background: white; border-radius: 12px; 
      box-shadow: 0 10px 30px rgba(0,0,0,0.2);
      overflow: hidden;
    }
    .header { 
      background: linear-gradient(135deg, #2c3e50, #34495e);
      color: white; padding: 30px; 
      border-bottom: 3px solid #3498db;
    }
    .header h1 { margin: 0 0 15px 0; font-size: 2.5em; }
    .header .meta { display: flex; gap: 30px; flex-wrap: wrap; }
    .header .meta-item { display: flex; flex-direction: column; }
    .header .meta-label { font-size: 0.9em; opacity: 0.8; margin-bottom: 5px; }
    .header .meta-value { font-size: 1.1em; font-weight: bold; }
    
    .content { padding: 30px; }
    .resource-grid { 
      display: grid; 
      grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); 
      gap: 20px; margin-bottom: 30px;
    }
    .resource-card { 
      border: 1px solid #e1e8ed; 
      border-radius: 8px; 
      overflow: hidden;
      transition: transform 0.2s, box-shadow 0.2s;
      cursor: pointer;
    }
    .resource-card:hover { 
      transform: translateY(-2px); 
      box-shadow: 0 5px 15px rgba(0,0,0,0.1);
    }
    .resource-header { 
      background: #f8f9fa; 
      padding: 15px; 
      border-bottom: 1px solid #e1e8ed;
    }
    .resource-title { 
      font-weight: bold; 
      font-size: 1.1em; 
      color: #2c3e50;
      margin-bottom: 5px;
    }
    .resource-meta { 
      font-size: 0.9em; 
      color: #6c757d;
      display: flex; gap: 15px; flex-wrap: wrap;
    }
    
    .summary-stats { 
      display: grid; 
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
      gap: 20px; margin-bottom: 30px;
    }
    .stat-card { 
      background: linear-gradient(135deg, #3498db, #2980b9);
      color: white; padding: 20px; 
      border-radius: 8px; text-align: center;
    }
    .stat-number { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
    .stat-label { font-size: 0.9em; opacity: 0.9; }
    
    .status-badge { 
      padding: 4px 8px; 
      border-radius: 12px; 
      font-size: 0.8em; 
      font-weight: bold;
    }
    .status-running { background: #d4edda; color: #155724; }
    .status-pending { background: #fff3cd; color: #856404; }
    .status-failed { background: #f8d7da; color: #721c24; }
    .status-unknown { background: #e2e3e5; color: #383d41; }
    
    .refresh-btn { 
      position: fixed; 
      bottom: 30px; right: 30px; 
      background: #3498db; color: white; 
      border: none; padding: 15px; 
      border-radius: 50%; 
      cursor: pointer; 
      box-shadow: 0 5px 15px rgba(0,0,0,0.2);
      transition: all 0.3s;
    }
    .refresh-btn:hover { 
      background: #2980b9; 
      transform: scale(1.1);
    }
    
    /* 模态框样式 */
    .modal {
      display: none;
      position: fixed;
      z-index: 1000;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      background-color: rgba(0,0,0,0.5);
      animation: fadeIn 0.3s;
      overflow-y: auto;
    }
    
    /* 阻止背景滚动 */
    body.modal-open {
      overflow: hidden;
    }
    
    /* 全屏模态框 */
    .modal.fullscreen .modal-content {
      width: 100vw;
      height: 100vh;
      max-width: none;
      max-height: none;
      margin: 0;
      border-radius: 0;
      animation: expandToFullscreen 0.3s ease-out;
    }
    
    .modal.fullscreen .modal-body {
      max-height: none;
      height: calc(100vh - 80px);
    }
    
    @keyframes expandToFullscreen {
      from { 
        width: 90%;
        height: 80vh;
        margin: 5% auto;
        border-radius: 12px;
      }
      to { 
        width: 100vw;
        height: 100vh;
        margin: 0;
        border-radius: 0;
      }
    }
    
    .modal-content {
      background-color: white;
      margin: 5% auto;
      padding: 0;
      border-radius: 12px;
      width: 90%;
      max-width: 1000px;
      max-height: 80vh;
      overflow: hidden;
      box-shadow: 0 20px 60px rgba(0,0,0,0.3);
      animation: slideIn 0.3s;
      display: flex;
      flex-direction: column;
    }
    
    .modal-header {
      background: linear-gradient(135deg, #2c3e50, #34495e);
      color: white;
      padding: 20px 30px;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    .modal-title {
      font-size: 1.5em;
      font-weight: bold;
      margin: 0;
    }
    
    .modal-subtitle {
      font-size: 0.9em;
      opacity: 0.8;
      margin: 5px 0 0 0;
    }
    
    .modal-controls {
      display: flex;
      align-items: center;
      gap: 15px;
    }
    
    .modal-control-btn {
      color: white;
      font-size: 20px;
      cursor: pointer;
      transition: opacity 0.3s;
      padding: 5px;
      border-radius: 4px;
      background: rgba(255,255,255,0.1);
    }
    
    .modal-control-btn:hover {
      opacity: 0.7;
      background: rgba(255,255,255,0.2);
    }
    
    .close {
      color: white;
      font-size: 28px;
      font-weight: bold;
      cursor: pointer;
      transition: opacity 0.3s;
    }
    
    .close:hover {
      opacity: 0.7;
    }
    
    .modal-body {
      padding: 30px;
      flex: 1;
      overflow-y: auto;
      min-height: 0;
    }
    
    .yaml-content { 
      background: #f8f9fa; 
      padding: 20px; 
      border-radius: 8px; 
      font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
      font-size: 0.9em;
      line-height: 1.4;
      overflow-x: auto;
      white-space: pre;
      border: 1px solid #e1e8ed;
      margin: 0;
    }
    
    /* 结构化资源显示 */
    .resource-section {
      margin-bottom: 25px;
      border: 1px solid #e1e8ed;
      border-radius: 8px;
      overflow: hidden;
    }
    
    .section-header {
      background: linear-gradient(135deg, #f8f9fa, #e9ecef);
      padding: 12px 20px;
      border-bottom: 1px solid #e1e8ed;
      font-weight: bold;
      color: #2c3e50;
      display: flex;
      justify-content: space-between;
      align-items: center;
      cursor: pointer;
      transition: background 0.3s;
    }
    
    .section-header:hover {
      background: linear-gradient(135deg, #e9ecef, #dee2e6);
    }
    
    .section-header .toggle-icon {
      font-size: 0.8em;
      transition: transform 0.3s;
    }
    
    .section-header.collapsed .toggle-icon {
      transform: rotate(-90deg);
    }
    
    .section-content {
      padding: 20px;
      background: white;
      max-height: 400px;
      overflow-y: auto;
    }
    
    .section-content.collapsed {
      display: none;
    }
    
    .key-value-grid {
      display: grid;
      grid-template-columns: 200px 1fr;
      gap: 15px 20px;
      align-items: start;
    }
    
    .key-label {
      font-weight: 600;
      color: #495057;
      padding: 8px 0;
      border-bottom: 1px solid #f1f3f4;
    }
    
    .value-content {
      padding: 8px 0;
      border-bottom: 1px solid #f1f3f4;
      word-break: break-word;
    }
    
    .value-object {
      background: #f8f9fa;
      padding: 15px;
      border-radius: 6px;
      border-left: 4px solid #3498db;
    }
    
    .value-array {
      background: #fff3cd;
      padding: 15px;
      border-radius: 6px;
      border-left: 4px solid #ffc107;
    }
    
    .array-item {
      padding: 8px 12px;
      margin: 5px 0;
      background: white;
      border-radius: 4px;
      border: 1px solid #e9ecef;
    }
    
    .nested-object {
      margin-left: 20px;
      padding-left: 15px;
      border-left: 2px solid #e9ecef;
    }
    
    .value-string {
      color: #28a745;
      font-family: 'Consolas', monospace;
    }
    
    .value-number {
      color: #007bff;
      font-family: 'Consolas', monospace;
    }
    
    .value-boolean {
      color: #dc3545;
      font-family: 'Consolas', monospace;
      font-weight: bold;
    }
    
    .value-null {
      color: #6c757d;
      font-style: italic;
    }
    
    .tab-buttons {
      display: flex;
      background: #f8f9fa;
      border-bottom: 1px solid #dee2e6;
      margin: -30px -30px 20px -30px;
    }
    
    .tab-button {
      padding: 12px 20px;
      border: none;
      background: none;
      cursor: pointer;
      font-size: 0.9em;
      color: #6c757d;
      transition: all 0.3s;
      border-bottom: 3px solid transparent;
    }
    
    .tab-button:hover {
      background: #e9ecef;
      color: #495057;
    }
    
    .tab-button.active {
      background: white;
      color: #2c3e50;
      border-bottom-color: #3498db;
      font-weight: bold;
    }
    
    .tab-content {
      display: none;
    }
    
    .tab-content.active {
      display: block;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; }
      to { opacity: 1; }
    }
    
    @keyframes slideIn {
      from { transform: translateY(-50px); opacity: 0; }
      to { transform: translateY(0); opacity: 1; }
    }
    
    @media (max-width: 768px) {
      .resource-grid { grid-template-columns: 1fr; }
      .header .meta { flex-direction: column; gap: 15px; }
      .modal-content { 
        width: 95%; 
        margin: 5% auto;
        max-height: 85vh;
      }
      .modal-header { padding: 15px 20px; }
      .modal-body { padding: 15px; }
      .key-value-grid { 
        grid-template-columns: 1fr; 
        gap: 10px;
      }
      .key-label {
        background: #f8f9fa;
        padding: 8px 12px;
        border-radius: 4px;
        margin-bottom: 5px;
      }
      .tab-buttons {
        margin: -15px -15px 15px -15px;
      }
      .tab-button {
        padding: 10px 15px;
        font-size: 0.8em;
      }
      .modal-controls {
        gap: 10px;
      }
      .modal-control-btn {
        font-size: 18px;
        padding: 3px;
      }
      .close {
        font-size: 24px;
      }
      
      /* 移动端全屏优化 */
      .modal.fullscreen .modal-content {
        width: 100vw;
        height: 100vh;
        margin: 0;
        border-radius: 0;
      }
      
      .modal.fullscreen .modal-body {
        height: calc(100vh - 70px);
        padding: 10px;
      }
      
      .modal.fullscreen .modal-header {
        padding: 10px 15px;
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>🚀 Kubernetes 资源查看器</h1>
      <div class="meta">
        <div class="meta-item">
          <div class="meta-label">执行命令</div>
          <div class="meta-value">kubectl {{ .Command }}</div>
        </div>
        <div class="meta-item">
          <div class="meta-label">生成时间</div>
          <div class="meta-value">{{ .Timestamp }}</div>
        </div>
        <div class="meta-item">
          <div class="meta-label">资源总数</div>
          <div class="meta-value">{{ .TotalResources }}</div>
        </div>
        <div class="meta-item">
          <div class="meta-label">命名空间</div>
          <div class="meta-value">{{ .NamespaceCount }}</div>
        </div>
      </div>
    </div>
    
    <div class="content">
      <div class="summary-stats">
        {{ range .KindStats }}
        <div class="stat-card">
          <div class="stat-number">{{ .Count }}</div>
          <div class="stat-label">{{ .Kind }}</div>
        </div>
        {{ end }}
      </div>
      
      {{ if .Resources }}
      <h3>📋 资源列表 (点击查看详情)</h3>
      <div class="resource-grid">
        {{ range $index, $resource := .Resources }}
        <div class="resource-card" onclick="showResourceModal({{ $index }})">
          <div class="resource-header">
            <div class="resource-title">{{ .Name }}</div>
            <div class="resource-meta">
              <span>🏷️ {{ .Kind }}</span>
              {{ if .Namespace }}<span>📁 {{ .Namespace }}</span>{{ end }}
              <span>⏰ {{ .Age }}</span>
              <span class="status-badge status-{{ .Status }}">{{ .Status }}</span>
            </div>
          </div>
        </div>
        {{ end }}
      </div>
      {{ end }}
    </div>
  </div>
  
  <!-- 模态框 -->
  <div id="resourceModal" class="modal">
    <div class="modal-content">
      <div class="modal-header">
        <div>
          <div class="modal-title" id="modalTitle">资源详情</div>
          <div class="modal-subtitle" id="modalSubtitle">YAML 配置</div>
        </div>
        <div class="modal-controls">
          <span class="modal-control-btn" onclick="toggleFullscreen()" id="fullscreenBtn" title="放大到全屏 (F11)">🔍</span>
          <span class="close" onclick="closeModal()" title="关闭">&times;</span>
        </div>
      </div>
      <div class="modal-body">
        <div class="tab-buttons">
          <button class="tab-button active" onclick="switchTab('structured')">📋 结构化视图</button>
          <button class="tab-button" onclick="switchTab('yaml')">📄 YAML 源码</button>
        </div>
        
        <div id="structuredTab" class="tab-content active">
          <div id="structuredContent">加载中...</div>
        </div>
        
        <div id="yamlTab" class="tab-content">
          <pre class="yaml-content" id="modalYaml">加载中...</pre>
        </div>
      </div>
    </div>
  </div>
  
  <button class="refresh-btn" onclick="location.reload()" title="刷新页面">🔄</button>
  
  <script>
    // 资源数据
    const resources = {{ .ResourcesJSON }};
    
    function renderValue(value, key = '') {
      if (value === null || value === undefined) {
        return '<span class="value-null">null</span>';
      }
      
      if (typeof value === 'string') {
        return '<span class="value-string">"' + escapeHtml(value) + '"</span>';
      }
      
      if (typeof value === 'number') {
        return '<span class="value-number">' + value + '</span>';
      }
      
      if (typeof value === 'boolean') {
        return '<span class="value-boolean">' + value + '</span>';
      }
      
      if (Array.isArray(value)) {
        if (value.length === 0) {
          return '<span class="value-null">[]</span>';
        }
        
        let html = '<div class="value-array">';
        html += '<strong>📋 数组 (' + value.length + ' 项)</strong>';
        value.forEach((item, index) => {
          html += '<div class="array-item">';
          if (typeof item === 'object' && item !== null) {
            html += '<strong>🔸 [' + index + ']</strong><br>';
            html += renderValue(item);
          } else {
            html += '<strong>🔸 [' + index + ']</strong> ' + renderValue(item);
          }
          html += '</div>';
        });
        html += '</div>';
        return html;
      }
      
      if (typeof value === 'object') {
        const keys = Object.keys(value);
        if (keys.length === 0) {
          return '<span class="value-null">{}</span>';
        }
        
        let html = '<div class="value-object">';
        html += '<strong>📦 对象 (' + keys.length + ' 个字段)</strong>';
        html += '<div class="key-value-grid" style="margin-top: 10px;">';
        
        keys.forEach(k => {
          html += '<div class="key-label">🔑 ' + escapeHtml(k) + '</div>';
          html += '<div class="value-content">' + renderValue(value[k], k) + '</div>';
        });
        
        html += '</div></div>';
        return html;
      }
      
      return '<span class="value-string">' + escapeHtml(String(value)) + '</span>';
    }
    
    function escapeHtml(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    }
    
    function renderStructuredResource(parsedResource) {
      try {
        if (!parsedResource || typeof parsedResource !== 'object') {
          return '<p>无法解析资源结构</p>';
        }
        
        // 生成结构化 HTML
        let html = '';
        
        // 主要部分
        const sections = [
          { key: 'apiVersion', title: 'API 版本', icon: '🔖' },
          { key: 'kind', title: '资源类型', icon: '📦' },
          { key: 'metadata', title: '元数据', icon: '📋' },
          { key: 'spec', title: '规格配置', icon: '⚙️' },
          { key: 'status', title: '状态信息', icon: '📊' },
          { key: 'data', title: '数据', icon: '💾' },
          { key: 'stringData', title: '字符串数据', icon: '📝' },
          { key: 'rules', title: '规则', icon: '📜' },
          { key: 'subjects', title: '主体', icon: '👥' },
          { key: 'roleRef', title: '角色引用', icon: '🔗' }
        ];
        
        sections.forEach(section => {
          if (parsedResource[section.key] !== undefined) {
            html += '<div class="resource-section">';
            html += '<div class="section-header" onclick="toggleSection(this)">';
            html += '<span>' + section.icon + ' ' + section.title + '</span>';
            html += '<span class="toggle-icon">▼</span>';
            html += '</div>';
            html += '<div class="section-content">';
            html += renderValue(parsedResource[section.key]);
            html += '</div>';
            html += '</div>';
          }
        });
        
        // 其他字段
        const otherKeys = Object.keys(parsedResource).filter(key => 
          !sections.some(section => section.key === key)
        );
        
        if (otherKeys.length > 0) {
          html += '<div class="resource-section">';
          html += '<div class="section-header" onclick="toggleSection(this)">';
          html += '<span>🔧 其他字段</span>';
          html += '<span class="toggle-icon">▼</span>';
          html += '</div>';
          html += '<div class="section-content">';
          html += '<div class="key-value-grid">';
          otherKeys.forEach(key => {
            html += '<div class="key-label">' + escapeHtml(key) + '</div>';
            html += '<div class="value-content">' + renderValue(parsedResource[key]) + '</div>';
          });
          html += '</div>';
          html += '</div>';
          html += '</div>';
        }
        
        return html || '<p>无法解析资源结构</p>';
        
      } catch (error) {
        return '<p>解析错误: ' + escapeHtml(error.message) + '</p>';
      }
    }
    
    function toggleSection(header) {
      const content = header.nextElementSibling;
      const icon = header.querySelector('.toggle-icon');
      
      if (content.classList.contains('collapsed')) {
        content.classList.remove('collapsed');
        header.classList.remove('collapsed');
        icon.textContent = '▼';
      } else {
        content.classList.add('collapsed');
        header.classList.add('collapsed');
        icon.textContent = '▶';
      }
    }
    
    function switchTab(tabName) {
      // 隐藏所有标签页
      document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
      });
      
      // 移除所有按钮的活动状态
      document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
      });
      
      // 显示选中的标签页
      document.getElementById(tabName + 'Tab').classList.add('active');
      event.target.classList.add('active');
    }
    
    function showResourceModal(index) {
      const resource = resources[index];
      const modal = document.getElementById('resourceModal');
      const title = document.getElementById('modalTitle');
      const subtitle = document.getElementById('modalSubtitle');
      const yaml = document.getElementById('modalYaml');
      const structured = document.getElementById('structuredContent');
      
      title.textContent = resource.name || 'Unknown Resource';
      subtitle.textContent = resource.kind + (resource.namespace ? ' (' + resource.namespace + ')' : '') + ' - ' + resource.apiVersion;
      yaml.textContent = resource.yaml;
      
      // 生成结构化视图
      structured.innerHTML = renderStructuredResource(resource.parsed);
      
      // 重置到结构化视图
      document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
      document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
      document.getElementById('structuredTab').classList.add('active');
      document.querySelector('.tab-button').classList.add('active');
      
      // 阻止背景滚动
      document.body.classList.add('modal-open');
      modal.style.display = 'block';
    }
    
    function closeModal() {
      const modal = document.getElementById('resourceModal');
      modal.style.display = 'none';
      modal.classList.remove('fullscreen');
      // 恢复背景滚动
      document.body.classList.remove('modal-open');
      // 重置全屏按钮
      const fullscreenBtn = document.getElementById('fullscreenBtn');
      fullscreenBtn.textContent = '🔍';
      fullscreenBtn.title = '放大到全屏 (F11)';
    }
    
    function toggleFullscreen() {
      const modal = document.getElementById('resourceModal');
      const fullscreenBtn = document.getElementById('fullscreenBtn');
      
      if (modal.classList.contains('fullscreen')) {
        // 退出全屏
        modal.classList.remove('fullscreen');
        fullscreenBtn.textContent = '🔍';
        fullscreenBtn.title = '放大到全屏 (F11)';
      } else {
        // 进入全屏
        modal.classList.add('fullscreen');
        fullscreenBtn.textContent = '🔎';
        fullscreenBtn.title = '退出全屏 (F11)';
      }
    }
    
    // 点击模态框外部关闭
    window.onclick = function(event) {
      const modal = document.getElementById('resourceModal');
      if (event.target === modal) {
        closeModal();
      }
    }
    
    // 键盘快捷键
    document.addEventListener('keydown', function(event) {
      const modal = document.getElementById('resourceModal');
      
      if (event.key === 'Escape') {
        closeModal();
      } else if (event.key === 'F11' && modal.style.display === 'block') {
        event.preventDefault();
        toggleFullscreen();
      }
    });
    
    // 阻止模态框内容滚动事件冒泡
    document.addEventListener('DOMContentLoaded', function() {
      const modalContent = document.querySelector('.modal-content');
      if (modalContent) {
        modalContent.addEventListener('wheel', function(e) {
          e.stopPropagation();
        });
        
        modalContent.addEventListener('touchmove', function(e) {
          e.stopPropagation();
        });
      }
    });
  </script>
</body>
</html>
`

type KindStat struct {
	Kind  string
	Count int
}



type PageData struct {
	Command        string
	Timestamp      string
	TotalResources int
	NamespaceCount int
	Resources      []ResourceInfo
	KindStats      []KindStat
	ResourcesJSON  template.JS
}

// 解析资源状态
func getResourceStatus(resource K8sResource) string {
	if resource.Status == nil {
		return "unknown"
	}

	statusMap, ok := resource.Status.(map[string]interface{})
	if !ok {
		return "unknown"
	}

	// 检查不同类型资源的状态
	switch resource.Kind {
	case "Pod":
		if phase, exists := statusMap["phase"]; exists {
			switch phase {
			case "Running":
				return "running"
			case "Pending":
				return "pending"
			case "Failed", "Error":
				return "failed"
			default:
				return "unknown"
			}
		}
	case "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet":
		if conditions, exists := statusMap["conditions"]; exists {
			if condList, ok := conditions.([]interface{}); ok {
				for _, cond := range condList {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if condType, exists := condMap["type"]; exists && condType == "Available" {
							if status, exists := condMap["status"]; exists && status == "True" {
								return "running"
							}
						}
					}
				}
			}
		}
		return "pending"
	case "Service":
		return "running"
	case "ConfigMap", "Secret":
		return "running"
	default:
		// CRD 或其他资源类型
		if conditions, exists := statusMap["conditions"]; exists {
			if condList, ok := conditions.([]interface{}); ok {
				for _, cond := range condList {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if condType, exists := condMap["type"]; exists {
							if strings.Contains(strings.ToLower(condType.(string)), "ready") ||
								strings.Contains(strings.ToLower(condType.(string)), "available") {
								if status, exists := condMap["status"]; exists && status == "True" {
									return "running"
								}
							}
						}
					}
				}
			}
		}
		return "unknown"
	}
	return "unknown"
}

// 计算资源年龄
func calculateAge(creationTimestamp interface{}) string {
	if creationTimestamp == nil {
		return "unknown"
	}

	timeStr, ok := creationTimestamp.(string)
	if !ok {
		return "unknown"
	}

	createdTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "unknown"
	}

	duration := time.Since(createdTime)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// 解析 YAML 数据
func parseKubernetesYAML(yamlData string) ([]K8sResource, error) {
	var resources []K8sResource

	// 分割多文档 YAML
	docs := strings.Split(yamlData, "---")
	
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 首先尝试解析为 List 类型
		var list K8sList
		if err := yaml.Unmarshal([]byte(doc), &list); err == nil && list.Kind == "List" {
			resources = append(resources, list.Items...)
			continue
		}

		// 然后尝试解析为单个资源
		var resource K8sResource
		if err := yaml.Unmarshal([]byte(doc), &resource); err == nil && resource.Kind != "" {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// 生成资源信息
func generateResourceInfo(resources []K8sResource) []ResourceInfo {
	var infos []ResourceInfo

	for _, resource := range resources {
		info := ResourceInfo{
			Kind:       resource.Kind,
			APIVersion: resource.APIVersion,
			Status:     getResourceStatus(resource),
		}

		// 提取名称和命名空间
		if resource.Metadata != nil {
			if name, exists := resource.Metadata["name"]; exists {
				info.Name = fmt.Sprintf("%v", name)
			}
			if namespace, exists := resource.Metadata["namespace"]; exists {
				info.Namespace = fmt.Sprintf("%v", namespace)
			}
			if creationTimestamp, exists := resource.Metadata["creationTimestamp"]; exists {
				info.Age = calculateAge(creationTimestamp)
			}
		}

		// 生成该资源的 YAML
		yamlBytes, err := yaml.Marshal(resource)
		if err == nil {
			info.YAML = string(yamlBytes)
		} else {
			info.YAML = "# YAML 生成失败: " + err.Error()
		}

		// 解析为 map 供前端使用
		var parsed map[string]interface{}
		if err := yaml.Unmarshal(yamlBytes, &parsed); err == nil {
			info.Parsed = parsed
		} else {
			info.Parsed = map[string]interface{}{
				"error": "解析失败: " + err.Error(),
			}
		}

		infos = append(infos, info)
	}

	return infos
}



// 生成种类统计
func generateKindStats(resources []K8sResource) []KindStat {
	kindCounts := make(map[string]int)

	for _, resource := range resources {
		kindCounts[resource.Kind]++
	}

	var stats []KindStat
	for kind, count := range kindCounts {
		stats = append(stats, KindStat{Kind: kind, Count: count})
	}

	// 按种类名称排序
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Kind < stats[j].Kind
	})

	return stats
}

// 计算命名空间数量
func countNamespaces(resources []K8sResource) int {
	namespaces := make(map[string]bool)

	for _, resource := range resources {
		if resource.Metadata != nil {
			if namespace, exists := resource.Metadata["namespace"]; exists {
				namespaces[fmt.Sprintf("%v", namespace)] = true
			}
		}
	}

	if len(namespaces) == 0 {
		return 1 // 可能是集群级别资源
	}

	return len(namespaces)
}

func main() {
	// 定义命令行参数
	var (
		host = flag.String("host", "localhost", "服务器监听地址 (localhost, 0.0.0.0, 或具体IP)")
		port = flag.String("port", "8000", "服务器监听端口")
		help = flag.Bool("help", false, "显示帮助信息")
	)

	// 解析命令行参数
	flag.Parse()

	if *help {
		fmt.Println("kubectl-html - Kubernetes 资源可视化工具")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Println("  kubectl-html [选项] [kubectl参数...]")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  -host string    服务器监听地址 (默认: localhost)")
		fmt.Println("                  localhost - 仅本机访问")
		fmt.Println("                  0.0.0.0   - 允许外部访问")
		fmt.Println("                  具体IP    - 绑定到指定网卡")
		fmt.Println("  -port string    服务器监听端口 (默认: 8000)")
		fmt.Println("  -help           显示此帮助信息")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  kubectl-html get pods")
		fmt.Println("  kubectl-html -host 0.0.0.0 get pods")
		fmt.Println("  kubectl-html -host 0.0.0.0 -port 9000 get deployments -A")
		fmt.Println("  kubectl-html get po,svc,deploy -n kube-system")
		fmt.Println("")
		fmt.Println("安全提示:")
		fmt.Println("  使用 0.0.0.0 会允许网络中的其他设备访问")
		fmt.Println("  请确保网络环境安全，或使用防火墙限制访问")
		return
	}

	// 获取 kubectl 参数
	kubectlArgs := flag.Args()
	if len(kubectlArgs) == 0 {
		log.Fatal("错误: 需要提供 kubectl 参数\n\n" +
			"用法: kubectl-html [选项] [kubectl参数...]\n" +
			"示例: kubectl-html get pods\n" +
			"帮助: kubectl-html -help")
	}

	// 构造 kubectl 命令
	kubectlArgs = append(kubectlArgs, "-o", "yaml")

	cmd := exec.Command("kubectl", kubectlArgs...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	log.Printf("🚀 Running: kubectl %s", strings.Join(kubectlArgs, " "))
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ kubectl failed: %v\nStderr: %s", err, errBuf.String())
	}

	yamlData := outBuf.String()
	if yamlData == "" {
		log.Fatal("❌ No data returned from kubectl")
	}

	// 解析 Kubernetes 资源
	resources, err := parseKubernetesYAML(yamlData)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to parse YAML structure: %v", err)
		// 继续使用原始 YAML
	}

	log.Printf("📦 Parsed %d resources", len(resources))

	// 生成各种数据
	resourceInfos := generateResourceInfo(resources)
	kindStats := generateKindStats(resources)
	namespaceCount := countNamespaces(resources)

	// 将资源信息转换为 JSON 供前端使用
	resourcesJSON, err := json.Marshal(resourceInfos)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to marshal resources to JSON: %v", err)
		resourcesJSON = []byte("[]")
	}

	// 构造页面数据
	data := PageData{
		Command:        strings.Join(os.Args[2:], " "),
		Timestamp:      time.Now().Format("2006-01-02 15:04:05 MST"),
		TotalResources: len(resources),
		NamespaceCount: namespaceCount,
		Resources:      resourceInfos,
		KindStats:      kindStats,
		ResourcesJSON:  template.JS(resourcesJSON),
	}

	// 启动 HTTP 服务器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			log.Printf("❌ Template error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("❌ Template execution error: %v", err)
		}
	})

	// 添加 API 端点用于获取 JSON 数据
	http.HandleFunc("/api/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// 构造监听地址
	listenAddr := *host + ":" + *port
	
	fmt.Printf("\n✅ Kubernetes 资源查看器已启动!\n")
	
	// 显示访问地址
	if *host == "0.0.0.0" {
		fmt.Printf("🌐 Web界面: \n")
		fmt.Printf("   本机访问: http://localhost:%s\n", *port)
		fmt.Printf("   网络访问: http://<你的IP>:%s\n", *port)
		fmt.Printf("⚠️  警告: 允许外部网络访问，请确保网络安全!\n")
	} else if *host == "localhost" || *host == "127.0.0.1" {
		fmt.Printf("🌐 Web界面: http://localhost:%s\n", *port)
	} else {
		fmt.Printf("🌐 Web界面: http://%s:%s\n", *host, *port)
	}
	
	fmt.Printf("📦 资源总数: %d\n", len(resources))
	fmt.Printf("🏷️  资源类型: %d\n", len(kindStats))
	fmt.Printf("📁 命名空间: %d\n", namespaceCount)
	fmt.Printf("🎯 监听地址: %s\n", listenAddr)
	fmt.Printf("\n按 Ctrl+C 退出\n\n")
	
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}