import { Bubble } from "@/Components/Message/Message.styles";


export const Message = ({ content, isSender, pending }) => {
    return(
        <div className={`flex my-2 mx-0 ${isSender ? "justify-end" : "justify-start"}`}>
            <Bubble
                pending={pending} 
                className={`text-white ${isSender ? "bg-green-600" : "bg-red-600"}`}>
                <p>{content}</p>
            </Bubble>
        </div>
    )
};