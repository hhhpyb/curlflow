<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { 
  NModal, 
  NCard, 
  NForm, 
  NFormItem, 
  NInput, 
  NAutoComplete, 
  NButton, 
  NSpace 
} from 'naive-ui';

// 定义接口元数据结构
interface MetaData {
  summary?: string;
  description?: string;
  tags?: string[] | null;
  [key: string]: any;
}

const props = defineProps<{
  show: boolean;
  initialData: MetaData | null;
  existingTags: string[];
  currentFileName?: string;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'save', value: { summary: string; tag: string; description: string }): void;
}>();

// 本地表单状态
const formValue = ref({
  summary: '',
  tag: '',
  description: ''
});

// 监听初始数据和显示状态，同步到本地表单
watch(
  () => props.show,
  (isShowing) => {
    if (isShowing) {
      if (props.initialData) {
        formValue.value = {
          summary: props.initialData.summary || '',
          // 处理 tags 数组，只取第一个作为文件夹名
          tag: (props.initialData.tags && props.initialData.tags.length > 0) 
            ? props.initialData.tags[0] 
            : '',
          description: props.initialData.description || ''
        };
      } else {
        // Reset for new request
        formValue.value = {
          summary: '',
          tag: '',
          description: ''
        };
      }
    }
  },
  { immediate: true }
);

// 格式化 AutoComplete 选项
const autoCompleteOptions = computed(() => {
  return props.existingTags.map(t => ({
    label: t,
    value: t
  }));
});

const handleSave = () => {
  emit('save', {
    summary: formValue.value.summary.trim(),
    tag: formValue.value.tag.trim(),
    description: formValue.value.description.trim()
  });
  closeModal();
};

const closeModal = () => {
  emit('update:show', false);
};
</script>

<template>
  <n-modal
    :show="props.show"
    @update:show="val => emit('update:show', val)"
    preset="card"
    style="width: 500px"
    title="Request Settings / 接口设置"
    :bordered="false"
    size="huge"
    role="dialog"
    aria-modal="true"
  >
    <n-form label-placement="top">
      <n-form-item label="Name (Summary)">
        <n-input 
          v-model:value="formValue.summary" 
          placeholder="请输入接口名称，例如：获取用户信息" 
          @keyup.enter="handleSave"
        />
      </n-form-item>

      <n-form-item label="Folder (Tag)">
        <n-auto-complete
          v-model:value="formValue.tag"
          :options="autoCompleteOptions"
          placeholder="选择现有文件夹或输入新文件夹名称"
          clearable
        />
      </n-form-item>

      <n-form-item label="Description">
        <n-input
          v-model:value="formValue.description"
          type="textarea"
          placeholder="请输入接口的详细描述..."
          :autosize="{
            minRows: 3,
            maxRows: 6
          }"
        />
      </n-form-item>
    </n-form>

    <template #footer>
      <n-space justify="end">
        <n-button @click="closeModal">Cancel</n-button>
        <n-button type="primary" @click="handleSave">Save</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<style scoped>
/* 针对深色模式的微调（如果需要） */
:deep(.n-card-header__title) {
  font-weight: 600;
}
</style>
