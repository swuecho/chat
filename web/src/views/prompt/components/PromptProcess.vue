<template>
  <div>
    <n-tree :data="treeData" block-line :render-label="renderLabel" :render-suffix="renderSuffix"
      :expanded-keys="expandedKeys" @update:expanded-keys="handleExpand" />
    <n-button @click="addRootNode" style="margin-top: 10px">Add Step</n-button>
  </div>
</template>

<script setup>
import { ref, computed, watch, h } from 'vue'
import { NTree, NInput, NButton, NSpace } from 'naive-ui'
import { SvgIcon } from '@/components/common'


const props = defineProps(['value'])
const emit = defineEmits(['update:value'])

const treeData = computed(() => {
  return props.value.map((item, index) => createTreeNode(item, `${index}`))
})

const expandedKeys = ref([])

const createTreeNode = (item, key) => {
  return {
    key,
    description: item.description,
    children: item.children ? item.children.map((child, childIndex) => createTreeNode(child, `${key}-${childIndex}`)) : undefined
  }
}

const handleExpand = (keys) => {
  expandedKeys.value = keys
}

const renderLabel = (info) => {
  const { option } = info
  return h(NInput, {
    value: option.description,
    onUpdateValue: (value) => {
      updateNodeValue(option.key, value)
    }
  })
}

const renderSuffix = (info) => {
  const { option } = info
  return h(NSpace, null, {
    default: () => [
      h(NButton, { onClick: () => addChild(option.key), text: true }, { default: () => h(SvgIcon, { icon: 'mdi-plus' }) }),
      h(NButton, { onClick: () => addSibling(option.key), text: true }, { default: () => h(SvgIcon, { icon: 'mdi-account-plus' }) }),
      h(NButton, { onClick: () => removeNode(option.key), text: true }, { default: () => h(SvgIcon, { icon: 'mdi-delete' }) })
    ]
  })
}

const updateNodeValue = (key, value) => {
  const newValue = updateTreeData(props.value, key.split('-'), (node) => {
    node.description = value
    return node
  })
  emit('update:value', newValue)
}

const addChild = (key) => {
  const newValue = updateTreeData(props.value, key.split('-'), (node) => {
    if (!node.children) node.children = []
    node.children.push({ description: 'New Child Step' })
    return node
  })
  emit('update:value', newValue)
  expandedKeys.value = [...expandedKeys.value, key]
}

const addSibling = (key) => {
  const keyParts = key.split('-')
  if (keyParts.length === 1) {
    // Adding a sibling to a root node
    const newValue = [...props.value, { description: 'New Root Step' }]
    emit('update:value', newValue)
  } else {
    const parentKeyParts = keyParts.slice(0, -1)
    const newValue = updateTreeData(props.value, parentKeyParts, (node) => {
      if (!node.children) node.children = []
      node.children.push({ description: 'Next Step' })
      return node
    })
    emit('update:value', newValue)
    expandedKeys.value = [...expandedKeys.value, parentKeyParts.join('-')]
  }
}

const removeNode = (key) => {
  const keyParts = key.split('-')
  if (keyParts.length === 1) {
    // Removing a root node
    const index = parseInt(keyParts[0])
    const newValue = [...props.value]
    newValue.splice(index, 1)
    emit('update:value', newValue)
  } else {
    const newValue = updateTreeData(props.value, keyParts.slice(0, -1), (node) => {
      const index = parseInt(keyParts[keyParts.length - 1])
      node.children.splice(index, 1)
      if (node.children.length === 0) delete node.children
      return node
    })
    emit('update:value', newValue)
  }
}

const addRootNode = () => {
  const newValue = [...props.value, { description: 'New Root Step' }]
  emit('update:value', newValue)
}

const updateTreeData = (data, keyParts, updateFn) => {
  if (keyParts.length === 1) {
    const index = parseInt(keyParts[0])
    const newData = [...data]
    newData[index] = updateFn(newData[index])
    return newData
  } else {
    const index = parseInt(keyParts[0])
    const newData = [...data]
    newData[index] = {
      ...newData[index],
      children: updateTreeData(newData[index].children, keyParts.slice(1), updateFn)
    }
    return newData
  }
}

watch(() => props.value, () => {
  // Update treeData when props.value changes
}, { deep: true })
</script>