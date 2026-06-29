import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { FiEdit2, FiPlus, FiTrash2 } from "react-icons/fi"
import { createProduct, deleteProduct, getAllProducts, updateProduct } from "@/api/product"
import type { Product } from "@/api/product/types"
import { queryKeys } from "@/api/queryKeys"
import AddProductAlert from "@/components/admin/dialogs/AddProductAlert"
import EditProductSheet from "@/components/admin/dialogs/EditProductSheet"

function ProductsTab() {
	const { data: products, isLoading } = useQuery({
		queryKey: queryKeys.products,
		queryFn: getAllProducts,
	})

	const [isEditModalOpen, setIsEditModalOpen] = useState(false)
	const [productToEdit, setProductToEdit] = useState<Product | null>(null)
	const [newProductName, setNewProductName] = useState("")
	const [isAddModalOpen, setIsAddModalOpen] = useState(false)

	const queryClient = useQueryClient()
	const { mutate: editProductMutation } = useMutation({
		mutationFn: updateProduct,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.products })
			setIsEditModalOpen(false)
			setProductToEdit(null)
			setNewProductName("")
		},
	})

	const { mutate: createProductMutation } = useMutation({
		mutationFn: createProduct,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.products })
		},
	})

	const { mutate: deleteProductMutation } = useMutation({
		mutationFn: deleteProduct,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.products })
		},
	})

	const handleAddProduct = () => {
		setIsAddModalOpen(true)
	}

	const handleEditProduct = (product: Product) => {
		setProductToEdit(product)
		setNewProductName(product.name)
		setIsEditModalOpen(true)
	}

	const handleConfirmEdit = () => {
		if (productToEdit && newProductName.trim()) {
			editProductMutation({
				id: productToEdit.id,
				name: newProductName.trim(),
			})
		}
	}

	const handleCloseEditModal = () => {
		setIsEditModalOpen(false)
		setProductToEdit(null)
		setNewProductName("")
	}

	const handleDeleteProduct = (product: Product) => {
		deleteProductMutation(product.id.toString())
	}

	const handleConfirmAdd = () => {
		if (newProductName.trim()) {
			createProductMutation({ name: newProductName.trim() })
			handleCloseAddModal()
		}
	}

	const handleCloseAddModal = () => {
		setIsAddModalOpen(false)
		setNewProductName("")
	}

	if (isLoading) {
		return (
			<div className="flex justify-center py-8">
				<span className="loading loading-spinner loading-md"></span>
			</div>
		)
	}

	return (
		<div className="p-2 space-y-4 h-full">
			<button
				type="button"
				className="btn btn-primary btn-block gap-2 p-2"
				onClick={handleAddProduct}
			>
				<FiPlus className="w-4 h-4" />
				Добавить продукт
			</button>

			<div className="space-y-2 overflow-y-auto p-2">
				{products?.length === 0 ? (
					<div className="text-center py-8 text-base-content/60">Здесь пока пусто</div>
				) : (
					products?.map((product: Product) => (
						<div key={product.id} className="card bg-base-100 shadow-sm">
							<div className="card-body p-4">
								<div className="flex justify-between items-start">
									<div className="flex-1">
										<h3 className="font-medium text-base-content">{product.name}</h3>
										<p className="text-sm text-base-content/60 mt-1">
											{new Date(product.created_at).toLocaleTimeString("ru-RU", {
												day: "2-digit",
												month: "2-digit",
												year: "numeric",
												hour: "2-digit",
												minute: "2-digit",
												second: "2-digit",
											})}
										</p>
									</div>
									<div className="flex gap-2 ml-4">
										<button
											type="button"
											className="btn btn-ghost btn-sm btn-square"
											onClick={() => handleEditProduct(product)}
										>
											<FiEdit2 className="w-4 h-4" />
										</button>
										<button
											type="button"
											className="btn btn-ghost btn-sm btn-square text-error"
											onClick={() => handleDeleteProduct(product)}
										>
											<FiTrash2 className="w-4 h-4" />
										</button>
									</div>
								</div>
							</div>
						</div>
					))
				)}
			</div>

			<EditProductSheet
				isOpen={isEditModalOpen}
				onClose={handleCloseEditModal}
				onConfirm={handleConfirmEdit}
				newProductName={newProductName}
				setNewProductName={setNewProductName}
			/>

			<AddProductAlert
				isOpen={isAddModalOpen}
				onClose={handleCloseAddModal}
				onConfirm={handleConfirmAdd}
				newProductName={newProductName}
				setNewProductName={setNewProductName}
			/>
		</div>
	)
}

export default ProductsTab
