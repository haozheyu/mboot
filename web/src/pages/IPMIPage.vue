<template>
  <div class="space-y-4">
    <section class="card p-5">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div><h1 class="text-lg font-semibold">设备控制器</h1><p class="mt-1 text-sm text-neutral-500">设备保存 DHCP/PXE 身份和动态地址；控制器提供电源及启动控制。当前已实现 IPMI/BMC。</p></div>
        <button class="btn btn-primary" @click="newNode">添加控制器</button>
      </div>
      <p v-if="message" class="mt-3 text-sm" :class="failed ? 'text-red-600':'text-neutral-600'">{{ message }}</p>
    </section>
    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
      <section class="card overflow-hidden">
        <table class="w-full text-sm"><thead class="border-b bg-neutral-50 text-left text-xs text-neutral-500"><tr><th class="px-4 py-3">控制器</th><th class="px-4 py-3">类型 / 地址</th><th class="px-4 py-3">关联设备</th><th class="px-4 py-3">当前业务 IP</th><th class="px-4 py-3">能力</th><th class="px-4 py-3 text-right">操作</th></tr></thead>
          <tbody class="divide-y"><tr v-for="node in nodes" :key="node.id"><td class="px-4 py-3 font-medium">{{ node.name }}</td><td class="px-4 py-3"><div>IPMI/BMC</div><div class="text-xs text-neutral-500">{{ node.address }}</div></td><td class="px-4 py-3"><template v-if="node.client"><div>{{ node.client.name }}</div><div class="text-xs text-neutral-500">{{ node.client.mac || '尚未识别 PXE MAC' }}</div></template><span v-else class="text-amber-600">未关联</span></td><td class="px-4 py-3"><span v-if="node.client?.ip">{{ node.client.ip }}</span><span v-else class="text-neutral-400">等待 DHCP/iPXE 上报</span></td><td class="px-4 py-3">BMC / Power / Boot<span v-if="node.vendor !== 'generic'"> / BIOS*</span></td><td class="px-4 py-3"><div class="flex justify-end gap-1"><button class="btn" @click="probe(node)">探测</button><button class="btn" @click="edit(node)">管理</button></div></td></tr></tbody>
        </table><div v-if="!nodes.length" class="p-10 text-center text-sm text-neutral-500">尚未添加设备控制器。</div>
      </section>
      <aside class="card p-4"><h2 class="font-medium">{{ form.id ? `管理 ${form.name}` : '添加控制器' }}</h2>
        <div class="mt-4 space-y-3">
          <div><label class="label">控制器类型</label><select class="input mt-1 w-full" value="ipmi" disabled><option value="ipmi">IPMI / BMC（已支持）</option></select><p class="mt-1 text-xs text-neutral-500">VMware、Proxmox、libvirt、Hyper-V 将作为其他控制器适配器接入，不要求虚拟机提供 BMC。</p></div>
          <div><label class="label">控制器名称</label><input v-model.trim="form.name" class="input mt-1 w-full"></div>
          <div><label class="label">BMC 地址</label><input v-model.trim="form.address" class="input mt-1 w-full" placeholder="192.168.1.20"></div>
          <div><label class="label">关联设备</label><select v-model.number="form.device_id" class="input mt-1 w-full"><option :value="0">暂不关联</option><option v-for="client in clients" :key="client.id" :value="client.id">{{ client.name }} · {{ client.mac || '待识别 MAC' }} · {{ client.ip || '暂无 IP' }}</option></select><p class="mt-1 text-xs leading-5 text-neutral-500">控制器按设备 ID 关联；设备再通过 PXE MAC 动态跟踪 DHCP 地址。物理机和虚拟机使用同一设备模型。</p></div>
          <div class="grid grid-cols-2 gap-2"><div><label class="label">用户名</label><input v-model.trim="form.username" class="input mt-1 w-full"></div><div><label class="label">密码</label><input v-model="form.password" type="password" class="input mt-1 w-full" :placeholder="form.id ? '留空保持不变':''"></div></div>
          <div class="grid grid-cols-2 gap-2"><div><label class="label">IPMI 接口</label><select v-model="form.interface" class="input mt-1 w-full"><option value="lanplus">lanplus（IPMI 2.0，推荐）</option><option value="lan">lan（IPMI 1.5，兼容旧设备）</option></select></div><div><label class="label">厂商</label><select v-model="form.vendor" class="input mt-1 w-full"><option value="generic">通用</option><option value="dell">Dell</option><option value="hpe">HPE</option><option value="lenovo">Lenovo</option><option value="inspur">浪潮</option><option value="supermicro">Supermicro</option></select></div></div>
          <p v-if="form.interface === 'lan'" class="rounded border border-amber-200 bg-amber-50 p-2 text-xs text-amber-700">lan 使用 IPMI 1.5/RMCP。若出现 “Authentication type NONE not supported”，请改用 lanplus。</p>
          <button class="btn btn-primary w-full" :disabled="busy" @click="save">保存</button>
          <template v-if="form.id">
          <div class="border-t pt-3">
            <div class="flex items-center justify-between"><label class="label">电源控制</label><button class="btn h-8" :disabled="busy" @click="power('status')">查询当前状态</button></div>
            <p class="mt-2 text-xs leading-5 text-neutral-500">以下按钮会直接操作物理服务器。软关机依赖操作系统响应；强制重启、断电和电源循环可能造成数据丢失。</p>
            <div class="mt-3 grid grid-cols-2 gap-2">
              <button v-for="item in powerActions" :key="item.action" class="btn justify-start text-left" :class="item.danger ? 'border-red-200 text-red-700 hover:bg-red-50' : ''" :disabled="busy" @click="power(item.action)">
                <span><span class="block text-sm font-medium">{{ item.label }}</span><span class="block text-[11px] font-normal text-neutral-500">{{ item.short }}</span></span>
              </button>
            </div>
          </div>
          <div class="border-t pt-3">
            <label class="label">设置下一次启动设备</label>
            <p class="mt-2 rounded border border-blue-200 bg-blue-50 p-2 text-xs leading-5 text-blue-700">这里只写入 BMC 的一次性启动覆盖，不会立即重启，也不会修改 BIOS 中的长期启动顺序。设置将在服务器下一次启动时生效，成功启动后通常自动失效。</p>
            <div class="mt-3 space-y-2">
              <select v-model="bootDevice" class="input w-full"><option value="pxe">PXE 网络启动（下一次）</option><option value="disk">本地硬盘启动（下一次）</option><option value="cdrom">虚拟光驱启动（下一次）</option><option value="bios">进入 BIOS Setup（下一次）</option></select>
              <select v-model="bootMode" class="input w-full"><option value="" disabled>请选择启动固件模式</option><option value="uefi">UEFI 模式</option><option value="legacy">Legacy BIOS 模式</option></select>
              <button class="btn w-full" :disabled="busy || !bootMode" @click="setBoot">确认并设置一次性启动项</button>
            </div>
            <p class="mt-2 text-xs text-neutral-500">设置完成后如需立刻执行，请再单独选择“软关机”“强制重启”或“电源循环”；系统不会自动执行电源操作。</p>
          </div>
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
type Client={id:number;name:string;ip:string;mac:string;status:string}
type Node={id:number;type?:'ipmi';device_id:number;client_id?:number;name:string;address:string;username:string;interface:string;vendor:string;has_password:boolean;created_at:string;updated_at:string;client?:Client}
const nodes=ref<Node[]>([]),clients=ref<Client[]>([]),busy=ref(false),message=ref(''),failed=ref(false),output=ref(''),bootDevice=ref('pxe'),bootMode=ref<''|'legacy'|'uefi'>('')
const powerActions = [
  { action: 'on', label: '开机', short: '服务器上电', danger: false },
  { action: 'soft', label: '软关机', short: '请求操作系统关机', danger: false },
  { action: 'reset', label: '强制重启', short: '类似按 Reset 键', danger: true },
  { action: 'cycle', label: '电源循环', short: '断电后重新上电', danger: true },
  { action: 'off', label: '强制断电', short: '立即切断电源', danger: true }
] as const
const powerDetails:Record<string,string>={on:'服务器将立即上电。',soft:'BMC 将发送软关机请求，需要操作系统支持 ACPI；不会等待业务安全退出。',reset:'服务器将被立即硬重启，未写入磁盘的数据可能丢失。',cycle:'服务器将先断电再重新上电，运行中的业务会中断，未写入数据可能丢失。',off:'服务器将立即断电，运行中的业务会中断，未写入数据可能丢失。'}
const bootLabels:Record<string,string>={pxe:'PXE 网络启动',disk:'本地硬盘启动',cdrom:'虚拟光驱启动',bios:'BIOS Setup'}
const form=reactive({...empty(),password:''})
function empty():Node{return {id:0,device_id:0,name:'',address:'',username:'ADMIN',interface:'lanplus',vendor:'generic',has_password:false,created_at:'',updated_at:''}}
async function load(){const [nodeRows,clientRows]=await Promise.all([api<Node[]>('/controllers'),api<Client[]>('/devices')]);nodes.value=nodeRows;clients.value=clientRows}
function edit(n:Node){Object.assign(form,n,{password:''});output.value=''} function newNode(){Object.assign(form,empty(),{password:''});output.value=''}
async function run(fn:()=>Promise<void>){busy.value=true;failed.value=false;try{await fn()}catch(e){failed.value=true;message.value=e instanceof Error?e.message:'操作失败'}finally{busy.value=false}}
async function save(){await run(async()=>{const path=form.id?`/controllers/${form.id}`:'/controllers';const saved=await api<Node>(path,{method:form.id?'PUT':'POST',body:JSON.stringify({...form,type:'ipmi'})});message.value='控制器已保存';await load();edit(saved)})}
async function probe(n:Node){edit(n);await run(async()=>{const r=await api<{bmc:string;power:string}>(`/controllers/${n.id}/probe`,{method:'POST'});output.value=`${r.bmc}\n\n${r.power}`})}
async function power(action:string){
  if(action!=='status'&&!confirm(`目标服务器：${form.name}（${form.address}）\n操作：${powerActions.find(item=>item.action===action)?.label??action}\n\n${powerDetails[action]}\n\n确认继续吗？`))return
  await run(async()=>{const r=await api<{output:string}>(`/controllers/${form.id}/power`,{method:'POST',body:JSON.stringify({action,confirm:action!=='status'})});output.value=r.output;message.value=action==='status'?`已查询 ${form.name} 的电源状态`:`${form.name} 的电源操作已发送`})
}
async function setBoot(){
  if(!bootMode.value)return
  const mode=bootMode.value==='uefi'?'UEFI':'Legacy BIOS'
  if(!confirm(`目标服务器：${form.name}（${form.address}）\n下一次启动：${bootLabels[bootDevice.value]}\n固件模式：${mode}\n有效范围：仅下一次启动\n自动重启：否\n\n此操作会覆盖 BMC 中已有的一次性启动设置。确认继续吗？`))return
  await run(async()=>{const r=await api<{output:string}>(`/controllers/${form.id}/boot`,{method:'POST',body:JSON.stringify({device:bootDevice.value,boot_mode:bootMode.value,persistent:false,power_action:'',confirm:true})});output.value=r.output;message.value=`${form.name} 已设置为下一次从${bootLabels[bootDevice.value]}（${mode}），未执行重启`})
}
async function bios(){await run(async()=>{const r=await api<{standard_boot_options:string;note:string}>(`/controllers/${form.id}/bios`);output.value=`${r.standard_boot_options}\n\n${r.note}`})}
async function remove(){if(!confirm(`确认删除控制器 ${form.name}？设备记录不会被删除。`))return;await run(async()=>{await api(`/controllers/${form.id}`,{method:'DELETE'});newNode();await load()})}
onMounted(()=>run(load))
</script>
