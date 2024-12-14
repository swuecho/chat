<template>
        <div class="container mx-auto p-4 flex">
                <div class="w-1/4 pr-4 border-r border-gray-200">
                        <div class="sticky top-0 z-10 bg-white">
                                <div v-if="stickyMonth" class="font-semibold text-lg mb-2 p-2 bg-gray-100 rounded">
                                        {{ stickyMonth }}
                                </div>
                        </div>
                        <ul class="mt-2">
                                <li v-for="yearMonth in Object.keys(groupedBlogs)" :key="yearMonth" class="py-2">
                                        <a :href="`#${yearMonth}`" class="block hover:text-blue-500">
                                                {{ yearMonth }}
                                        </a>
                                </li>
                        </ul>
                </div>
                <div class="w-3/4 overflow-y-auto h-[calc(100vh-100px)]" ref="blogContainer">
                        <div v-for="[yearMonth, blogs] in Object.entries(groupedBlogs)" :key="yearMonth" :id="yearMonth"
                                :data-year-month="yearMonth" class="mb-8">
                                <h2 class="text-xl font-semibold mb-4">{{ yearMonth }}</h2>
                                <ul>
                                        <li v-for="blog in blogs" :key="blog.id"
                                                class="mb-4 border border-gray-200 rounded p-4">
                                                <h3 class="font-bold text-gray-900">{{ blog.title }}</h3>
                                                <p class="text-gray-700">{{ blog.content }}</p>
                                        </li>
                                </ul>
                        </div>
                </div>
        </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, watch } from 'vue';

interface Blog {
        id: number;
        title: string;
        date: string; // Assuming date is in 'YYYY-MM-DD' format
        content: string;
}

interface GroupedBlogs {
        [yearMonth: string]: Blog[];
}

const groupedBlogs = ref<GroupedBlogs>({});
const stickyMonth = ref<string | null>(null);
const blogContainer = ref<HTMLElement | null>(null);
const observer = ref<IntersectionObserver | null>(null);

// Sample Blog Data (Replace with your API Call)
const blogData: Blog[] = [
        { id: 1, title: 'First Blog', date: '2023-10-15', content: 'Content 1' },
        { id: 2, title: 'Second Blog', date: '2023-10-20', content: 'Content 2' },
        { id: 3, title: 'Third Blog', date: '2023-11-05', content: 'Content 3' },
        { id: 4, title: 'Fourth Blog', date: '2023-11-10', content: 'Content 4' },
        { id: 5, title: 'Fifth Blog', date: '2023-12-01', content: 'Content 5' },
        { id: 6, title: 'Sixth Blog', date: '2024-01-01', content: 'Content 6' },
        { id: 7, title: 'Seventh Blog', date: '2024-01-10', content: 'Content 7' },
        { id: 8, title: 'Eighth Blog', date: '2024-02-15', content: 'Content 8' },
        { id: 9, title: 'Ninth Blog', date: '2024-02-20', content: 'Content 9' },
        { id: 10, title: 'Tenth Blog', date: '2024-03-05', content: 'Content 10' },
        // ... more blogs
];


// Group the blogs by year-month
const grouped = blogData.reduce((acc: GroupedBlogs, blog) => {
        const yearMonth = blog.date.substring(0, 7);
        if (!acc[yearMonth]) {
                acc[yearMonth] = [];
        }
        acc[yearMonth].push(blog);
        return acc;
}, {});
groupedBlogs.value = grouped;

onMounted(() => {
        const observerOptions = {
                root: blogContainer.value,
                rootMargin: '0px 0px -90% 0px',
                threshold: 0
        }
        observer.value = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                        if (entry.isIntersecting) {
                                stickyMonth.value = entry.target.dataset.yearMonth || null
                        }
                })
        }, observerOptions)

        const monthElements = blogContainer.value?.querySelectorAll('[data-year-month]');
        monthElements?.forEach((monthElement) => {
                observer.value?.observe(monthElement)
        })
});

onUnmounted(() => {
        observer.value?.disconnect();
});
</script>

<style scoped>
/* Add any scoped styles here if needed */
body {
        margin: 0;
}
</style>