import { createSignal } from "solid-js";
import { definePage } from "../runtime";

const meta = () => ({
  title: "Home",
});

const Home = (props) => {
  const [count, setCount] = createSignal(0);

  return (
    <div>
      <p>Count: {count()}</p>
      <button onClick={() => setCount((c) => c + 1)}>Incr</button>
      {JSON.stringify(props, null, 2)}
    </div>
  );
};

export default definePage(Home, { meta });
