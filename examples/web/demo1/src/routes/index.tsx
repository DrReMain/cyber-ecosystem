import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import {
	useBlogList,
	useCreateBlog,
	useDeleteBlog,
} from "../lib/hooks/use-blog.ts";
import { useReadingBlogList, useRecordReading } from "../lib/hooks/use-reading.ts";

export const Route = createFileRoute("/")({ component: App });

function App() {
	const [title, setTitle] = useState("");
	const [content, setContent] = useState("");

	const {
		data: blogList,
		isLoading: blogLoading,
		error: blogError,
		refetch: refetchBlogs,
	} = useBlogList({ pageNo: 1, pageSize: 10 });
	const createBlog = useCreateBlog();
	const deleteBlog = useDeleteBlog();

	const {
		data: readingList,
		isLoading: readingLoading,
		error: readingError,
	} = useReadingBlogList({ pageNo: 1, pageSize: 10 });
	const recordReading = useRecordReading();

	const handleCreateBlog = async () => {
		if (!title.trim()) return;
		await createBlog.mutateAsync({ title, content });
		setTitle("");
		setContent("");
		refetchBlogs();
	};

	return (
		<main className="page-wrap px-4 pb-8 pt-14">
			<h1 className="text-2xl font-bold mb-6">Demo1 - Connect-Web</h1>

			{/* 创建博客表单 */}
			<section className="mb-8 p-4 border rounded-lg">
				<h2 className="text-xl font-semibold mb-4">创建博客 (Template1)</h2>
				<div className="flex flex-col gap-3">
					<input
						type="text"
						placeholder="标题"
						value={title}
						onChange={(e) => setTitle(e.target.value)}
						className="border p-2 rounded"
					/>
					<textarea
						placeholder="内容"
						value={content}
						onChange={(e) => setContent(e.target.value)}
						className="border p-2 rounded"
						rows={3}
					/>
					<button
						onClick={handleCreateBlog}
						disabled={createBlog.isPending}
						className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
					>
						{createBlog.isPending ? "创建中..." : "创建博客"}
					</button>
				</div>
			</section>

			{/* Template1 - Blog 列表 */}
			<section className="mb-8">
				<h2 className="text-xl font-semibold mb-4">
					博客列表 (Template1 - BlogService)
				</h2>
				{blogLoading && <p>加载中...</p>}
				{blogError && <p className="text-red-500">错误: {blogError.message}</p>}
				{blogList && (
					<div className="space-y-3">
						{blogList.list?.length === 0 && (
							<p className="text-gray-500">暂无数据</p>
						)}
						{blogList.list?.map((blog) => (
							<div
								key={blog.id}
								className="p-4 border rounded-lg flex justify-between items-start"
							>
								<div>
									<h3 className="font-medium">{blog.title || "无标题"}</h3>
									<p className="text-gray-600 text-sm">
										{blog.content || "无内容"}
									</p>
									<p className="text-gray-400 text-xs mt-1">ID: {blog.id}</p>
								</div>
								<button
									onClick={() => deleteBlog.mutate({ id: blog.id })}
									disabled={deleteBlog.isPending}
									className="text-red-500 hover:text-red-700 text-sm"
								>
									删除
								</button>
							</div>
						))}
					</div>
				)}
			</section>

			{/* Template2 - Reading 列表 */}
			<section className="mb-8">
				<h2 className="text-xl font-semibold mb-4">
					阅读统计 (Template2 - ReadingService)
				</h2>
				{readingLoading && <p>加载中...</p>}
				{readingError && (
					<p className="text-red-500">错误: {readingError.message}</p>
				)}
				{readingList && (
					<div className="space-y-3">
						{readingList.list?.length === 0 && (
							<p className="text-gray-500">暂无数据</p>
						)}
						{readingList.list?.map((blog) => (
							<div
								key={blog.id}
								className="p-4 border rounded-lg flex justify-between items-center"
							>
								<div>
									<h3 className="font-medium">{blog.title || "无标题"}</h3>
									<p className="text-gray-600 text-sm">
										阅读次数: {blog.readingCount}
									</p>
								</div>
								<button
									onClick={() => recordReading.mutate({ id: blog.id })}
									disabled={recordReading.isPending}
									className="bg-green-500 text-white px-3 py-1 rounded text-sm hover:bg-green-600"
								>
									记录阅读
								</button>
							</div>
						))}
					</div>
				)}
			</section>
		</main>
	);
}
