<template>
  <div class="space-y-4">
    <section class="card p-5">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div><h1 class="text-lg font-semibold">IPMI 带外管理</h1><p class="mt-1 text-sm text-neutral-500">统一管理 BMC、Power、Boot；BIOS 能力按服务器厂商支持情况提供。</p></div>
        <button class="btn btn-primary" @click="newNode">添加 BMC</button>
      </div>
      <p v-if="message" class="mt-3 text-sm" :class="failed ? 'text-red-600':'text-neutral-600'">{{ message }}</p>
    </section>
    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
      <section class="card overflow-hidden">
        <table class="w-full text-sm"><thead class="border-b bg-neutral-50 text-left text-xs text-neutral-500"><tr><th class="px-4 py-3">节点</th><th class="px-4 py-3">BMC</th><th class="px-4 py-3">厂商</th><th class="px-4 py-3">能力</th><th class="px-4 py-3 text-right">操作</th></tr></thead>
          <tbody class="divide-y"><tr v-for="node in nodes" :key="node.id"><td class="px-4 py-3 font-medium">{{ node.name }}</td><td class="px-4 py-3">{{ node.address }}</td><td class="px-4 py-3">{{ node.vendor }}</td><td class="px-4 py-3">BMC / Power / Boot<span v-if="node.vendor !== 'generic'"> / BIOS*</span></td><td class="px-4 py-3"><div class="flex justify-end gap-1"><button class="btn" @click="probe(node)">探测</button><button class="btn" @click="edit(node)">管理</button></div></td></tr></tbody>
        </table><div v-if="!nodes.length" class="p-10 text-center text-sm text-neutral-500">尚未添加 BMC 节点。</div>
      </section>
      <aside class="card p-4"><h2 class="font-medium">{{ form.id ? `管理 ${form.name}` : '添加 BMC' }}</h2>
        <div class="mt-4 space-y-3">
          <div><label class="label">名称</label><input v-model.trim="form.name" class="input mt-1 w-full"></div>
          <div><label class="label">BMC 地址</label><input v-model.trim="form.address" class="input mt-1 w-full" placeholder="192.168.1.20"></div>
          <div class="grid grid-cols-2 gap-2"><div><label class="label">用户名</label><input v-model.trim="form.username" class="input mt-1 w-full"></div><div><label class="label">密码</label><input v-model="form.password" type="password" class="input mt-1 w-full" :placeholder="form.id ? '留空保持不变':''"></div></div>
          <div class="grid grid-cols-2 gap-2"><div><label class="label">IPMI 接口</label><select v-model="form.interface" class="input mt-1 w-full"><option>lanplus</option><option>lan</option></select></div><div><label class="label">厂商</label><select v-model="form.vendor" class="input mt-1 w-full"><option value="generic">通用</option><option value="dell">Dell</option><option value="hpe">HPE</option><option value="lenovo">Lenovo</option><option value="inspur">浪潮</option><option value="supermicro">Supermicro</option></select></div></div>
          <button class="btn btn-primary w-full" :disabled="busy" @click="save">保存</button>
          <template v-if="form.id"><div class="border-t pt-3"><label class="label">Power</label><div class="mt-2 flex flex-wrap gap-1"><button v-for="a in ['status','on','soft','cycle','reset','off']" :key="a" class="btn" :disabled="busy" @click="power(a)">{{ a }}</button></div></div>
          <div><label class="label">下一次启动</label><div class="mt-2 flex gap-2"><select v-model="bootDevice" class="input flex-1"><option value="pxe">PXE</option><option value="disk">本地磁盘</option><option value="cdrom">虚拟光驱</option><option value="bios">BIOS Setup</option></select><button class="btn" :disabled="busy" @click="setBoot">设置</button></div><label class="mt-2 flex items-center gap-2 text-xs"><input v-model="uefi" type="checkbox">UEFI 启动</label></div>
          <div><button class="btn" :disabled="busy" @click="bios">读取 BIOS 启动参数</button><p class="mt-2 text-xs text-neutral-500">* 完整 BIOS 属性并非 IPMI 标准能力，需要对应厂商 OEM 适配。</p></div>
          <button class="btn btn-danger w-full" :disabled="busy" @click="remove">删除节点</button></template>
          <pre v-if="output" class="max-h-72 overflow-auto whitespace-pre-wrap rounded bg-neutral-950 p-3 text-xs text-neutral-100">{{ output }}</pre>
        </div>
      </aside>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'; import { api } from '../lib/api'
type Node={id:number;client_id:number;name:string;address:string;username:string;interface:string;vendor:string;has_password:boolean;created_at:string;updated_at:string}
const nodes=ref<Node[]>([]),busy=ref(false),message=ref(''),failed=ref(false),output=ref(''),bootDevice=ref('pxe'),uefi=ref(true)
const form=reactive({...empty(),password:''})
function empty():Node{return {id:0,client_id:0,name:'',address:'',username:'ADMIN',interface:'lanplus',vendor:'generic',has_password:false,created_at:'',updated_at:''}}
async function load(){nodes.value=await api<Node[]>('/ipmi/nodes')}
function edit(n:Node){Object.assign(form,n,{password:''});output.value=''} function newNode(){Object.assign(form,empty(),{password:''});output.value=''}
async function run(fn:()=>Promise<void>){busy.value=true;failed.value=false;try{await fn()}catch(e){failed.value=true;message.value=e instanceof Error?e.message:'操作失败'}finally{busy.value=false}}
async function save(){await run(async()=>{const path=form.id?`/ipmi/nodes/${form.id}`:'/ipmi/nodes';const saved=await api<Node>(path,{method:form.id?'PUT':'POST',body:JSON.stringify(form)});message.value='节点已保存';await load();edit(saved)})}
async function probe(n:Node){edit(n);await run(async()=>{const r=await api<{bmc:string;power:string}>(`/ipmi/nodes/${n.id}/probe`,{method:'POST'});output.value=`${r.bmc}\n\n${r.power}`})}
async function power(action:string){if(['off','reset','cycle'].includes(action)&&!confirm(`确认对 ${form.name} 执行 power ${action}？`))return;await run(async()=>{const r=await api<{output:string}>(`/ipmi/nodes/${form.id}/power`,{method:'POST',body:JSON.stringify({action})});output.value=r.output})}
async function setBoot(){await run(async()=>{const r=await api<{output:string}>(`/ipmi/nodes/${form.id}/boot`,{method:'POST',body:JSON.stringify({device:bootDevice.value,uefi:uefi.value,persistent:false})});output.value=r.output;message.value='一次性启动项已设置'})}
async function bios(){await run(async()=>{const r=await api<{standard_boot_options:string;note:string}>(`/ipmi/nodes/${form.id}/bios`);output.value=`${r.standard_boot_options}\n\n${r.note}`})}
async function remove(){if(!confirm(`确认删除 ${form.name}？`))return;await run(async()=>{await api(`/ipmi/nodes/${form.id}`,{method:'DELETE'});newNode();await load()})}
onMounted(()=>run(load))
</script>
